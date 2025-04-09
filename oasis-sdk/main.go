package oasis_sdk

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"mellium.im/sasl"
	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

type LoginInfo struct {
	Host        string   `json:"Host"`
	User        string   `json:"User"`
	Password    string   `json:"Password"`
	DisplayName string   `json:"DisplayName"`
	TLSoff      bool     `json:"NoTLS"`
	StartTLS    bool     `json:"StartTLS"`
	MucsToJoin  []string `json:"Mucs"`
}

//XmppAbstractMessage struct is a representation of the stanza such that it's contextual items
//such as room, as well as abstract methods such as .reply()
type XmppAbstractMessage struct {
	Stanza struct {
		stanza.Message
		Body string `xml:"body"`
	}
}

type XmppMessageListener struct {
	StanzaType string
	MessageType stanza.MessageType
	BareJID jid.JID
	Resourcepart string
	EventChan chan
}

}

type XmppClient struct {
	Ctx       context.Context
	CtxCancel context.CancelFunc
	Login     *LoginInfo
	JID       *jid.JID
	Server    string
	Session   *xmpp.Session
}

//addHandler is an internal function to add an internal handler to the session
func (self *XmppClient) addHandler() {
	self.Session.Serve(
		xmpp.HandlerFunc(
			func(tokenReadEncoder xmlstream.TokenReadEncoder, start *xml.StartElement) error {

			})
		)
	)
}

func (self *XmppClient) Connect(blocking bool) error {
	d := dial.Dialer{}

	conn, err := d.DialServer(self.Ctx, "tcp", *self.JID, self.Server)
	if err != nil {
		return errors.New("Could not connect stage 1 - " + err.Error())
	}

	self.Session, err = xmpp.NewSession(self.Ctx, j.Domain(), j, conn, 0, xmpp.NewNegotiator(func(*xmpp.Session, *xmpp.StreamConfig) xmpp.StreamConfig {
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
	}))
	if err != nil {
		return errors.New("Could not connect stage 2 - " + err.Error())
	}

	if blocking {
		self.addHandler()
	} else {
		go self.addHandler()
	}

	return nil
}

func CreateClient(login LoginInfo) (XmppClient, error) {
	client := &XmppClient{}
	client.Ctx, client.CtxCancel = context.WithCancel(context.Background())

	j, err := jid.Parse(login.User)
	if err != nil {
		return *client,
			errors.New("Could not parse user JID from `" + login.User + " - " + err.Error())
	}

	client.Server = j.Domainpart()

	return *client, nil
}
