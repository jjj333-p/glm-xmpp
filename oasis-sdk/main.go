package oasis_sdk

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mellium.im/sasl"
	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
	"sync"
)

// LoginInfo is a struct of the information required to log into the xmpp  client
type LoginInfo struct {
	Host        string   `json:"Host"`
	User        string   `json:"User"`
	Password    string   `json:"Password"`
	DisplayName string   `json:"DisplayName"`
	TLSoff      bool     `json:"NoTLS"`
	StartTLS    bool     `json:"StartTLS"`
	MucsToJoin  []string `json:"Mucs"`
}

// XmppMessageBody is a struct representing a raw event stanza
type XmppMessageBody struct {
	stanza.Message
	Body string `xml:"body"`
}

/*
XmppAbstractMessage struct is a representation of the stanza such that it's contextual items
such as room, as well as abstract methods such as .reply()
*/
type XmppAbstractMessage struct {
	Stanza struct {
		stanza.Message
		Body string `xml:"body"`
	}
}

// xmppMessageListener contains internal metadata for event listener channels
type xmppMessageListener struct {
	StanzaType    string
	MessageType   stanza.MessageType
	BareJID       string
	Resourcepart  string
	LeftToRecieve int
	SwallowEvent  bool
	EventChan     chan XmppAbstractMessage
}

// xmppMessageListeners allows for thread safe accessing of listeners
type xmppMessageListeners struct {
	Array []*xmppMessageListener
	Lock  sync.Mutex
}

// XmppClient is the end xmpp client object from which everything else works around
type XmppClient struct {
	Ctx       context.Context
	CtxCancel context.CancelFunc
	Login     *LoginInfo
	JID       *jid.JID
	Server    string
	Session   *xmpp.Session
	listeners *xmppMessageListeners
}

// startServing is an internal function to add an internal handler to the session.
// Most of this is just obtuse things inherited from mellium
func (self *XmppClient) startServing() error {
	return self.Session.Serve(
		xmpp.HandlerFunc(
			func(tokenReadEncoder xmlstream.TokenReadEncoder, start *xml.StartElement) error {
				decoder := xml.NewTokenDecoder(xmlstream.MultiReader(xmlstream.Token(*start), tokenReadEncoder))
				if _, err := decoder.Token(); err != nil {
					return err
				}

				body := XmppMessageBody{}
				err := decoder.DecodeElement(&body, start)
				if err != nil && err != io.EOF {
					fmt.Println("Error decoding element - " + err.Error())
					return nil
				}

				self.listeners.Lock.Lock()

				indexesToRemove := make([]int, 0)
				//emit to every listener
				for i := len(self.listeners.Array) - 1; i >= 0; i-- {
					listener := self.listeners.Array[i]

					//check the conditionals
					if listener.StanzaType != "" && listener.StanzaType != start.Name.Local {
						continue
					}
					if listener.MessageType != "" && listener.MessageType != body.Type {
						continue
					}
					if listener.BareJID != "" && listener.BareJID != body.From.Bare().String() {
						continue
					}
					if listener.Resourcepart != "" && listener.Resourcepart != body.From.Resourcepart() {
						continue
					}
					if listener.LeftToRecieve == 0 {
						indexesToRemove = append(indexesToRemove, i)
						close(listener.EventChan)
					} else if listener.LeftToRecieve > 0 {
						listener.LeftToRecieve--
					}

					//emit event
					listener.EventChan <- XmppAbstractMessage{
						Stanza: body,
					}

					//latest consumers first, can swallow
					if listener.SwallowEvent {
						break
					}

				}
				self.listeners.Lock.Unlock()
				return nil
			},
		),
	)
}

type connectionErrHandler func(err error)

/*
Connect dials the server and starts recieving the events.
if blocking is true, this method will not exit until the xmpp connection is no longer being maintained
if blocking is false, this method will exit as soon as a connection is created, and errors will be emitted
through the callback onErr
*/
func (self *XmppClient) Connect(blocking bool, onErr connectionErrHandler) error {
	d := dial.Dialer{}

	conn, err := d.DialServer(self.Ctx, "tcp", *self.JID, self.Server)
	if err != nil {
		return errors.New("Could not connect stage 1 - " + err.Error())
	}

	self.Session, err = xmpp.NewSession(
		self.Ctx,
		self.JID.Domain(),
		*self.JID,
		conn,
		0,
		xmpp.NewNegotiator(func(*xmpp.Session, *xmpp.StreamConfig) xmpp.StreamConfig {
			return xmpp.StreamConfig{
				Lang: "en",
				Features: []xmpp.StreamFeature{
					xmpp.BindResource(),
					xmpp.StartTLS(&tls.Config{
						ServerName: self.Server,
						MinVersion: tls.VersionTLS12,
					}),
					xmpp.SASL("", self.Login.Password, sasl.ScramSha1Plus, sasl.ScramSha1, sasl.Plain),
				},
				TeeIn:  nil,
				TeeOut: nil,
			}
		},
		))
	if err != nil {
		return errors.New("Could not connect stage 2 - " + err.Error())
	}

	if blocking {
		return self.startServing()
	} else {
		//serve in a thread
		go func() {
			err := self.startServing()

			//if error try to callback error handler, otherwise panic
			if err != nil {
				if onErr == nil {
					panic(err)
				} else {
					onErr(err)
				}
			}
		}()
	}

	return nil
}

func (self *XmppClient) CreateListener(
	stanzaType string,
	messageType stanza.MessageType,
	bareJID string,
	resourcepart string,
	limit int,
	swallowEvent bool,
) chan XmppAbstractMessage {
	ch := make(chan XmppAbstractMessage)
	self.listeners.Lock.Lock()
	defer self.listeners.Lock.Unlock()
	self.listeners.Array = append(self.listeners.Array, &xmppMessageListener{
		StanzaType:    stanzaType,
		MessageType:   messageType,
		BareJID:       bareJID,
		Resourcepart:  resourcepart,
		LeftToRecieve: limit,
		SwallowEvent:  swallowEvent,
		EventChan:     ch,
	})
	return ch
}

// CreateClient creates the client object using the login info object, and returns it
func CreateClient(login LoginInfo) (XmppClient, error) {
	// create client object
	client := &XmppClient{}
	client.Ctx, client.CtxCancel = context.WithCancel(context.Background())

	//string to jid object
	j, err := jid.Parse(login.User)
	if err != nil {
		return *client,
			errors.New("Could not parse user JID from `" + login.User + " - " + err.Error())
	}

	client.Server = j.Domainpart()

	return *client, nil
}
