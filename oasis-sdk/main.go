package oasis_sdk

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"mellium.im/sasl"
	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"mellium.im/xmpp/mux"
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
	Ctx         context.Context
	CtxCancel   context.CancelFunc
	Login       *LoginInfo
	JID         *jid.JID
	Server      *string
	Session     *xmpp.Session
	listeners   *xmppMessageListeners
	Multiplexer *mux.ServeMux
	MucClient   *muc.Client
}

// startServing is an internal function to add an internal handler to the session.
// Most of this is just obtuse things inherited from mellium
func (self *XmppClient) startServing() error {
	err := self.Session.Send(self.Ctx, stanza.Presence{Type: stanza.AvailablePresence}.Wrap(nil))
	if err != nil {
		return err
	}
	return self.Session.Serve(
		self.Multiplexer,
	)
}

func (self *XmppClient) HandleDM(msg stanza.Message, t xmlstream.TokenReadEncoder) error {
	fmt.Println("bing bong")
	return nil
}

type connectionErrHandler func(err error)

/*
Connect dials the server and starts recieving the events.
If blocking is true, this method will not exit until the xmpp connection is no longer being maintained.
If blocking is false, this method will exit as soon as a connection is created, and errors will be emitted
through the callback onErr
*/
func (self *XmppClient) Connect(blocking bool, onErr connectionErrHandler) error {
	d := dial.Dialer{}

	conn, err := d.DialServer(self.Ctx, "tcp", *self.JID, *self.Server)
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
						ServerName: *self.Server,
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

//func (self *XmppClient) sendMessageToJID() {
//
//}

// CreateClient creates the client object using the login info object, and returns it
func CreateClient(login *LoginInfo) (XmppClient, error) {
	// create client object
	client := &XmppClient{}
	client.Ctx, client.CtxCancel = context.WithCancel(context.Background())
	client.Login = login

	//client.MucClient
	messageNS := xml.Name{
		//Space: "jabber:client",
		Local: "body",
	}

	client.Multiplexer = mux.New(
		"jabber:client",
		muc.HandleClient(client.MucClient),
		mux.MessageFunc(stanza.ChatMessage, messageNS, client.HandleDM),
	)

	//string to jid object
	j, err := jid.Parse(login.User)
	if err != nil {
		return *client,
			errors.New("Could not parse user JID from `" + login.User + " - " + err.Error())
	}
	server := j.Domainpart()
	client.JID = &j
	client.Server = &server

	return *client, nil
}
