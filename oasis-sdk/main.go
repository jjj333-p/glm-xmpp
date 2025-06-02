package oasis_sdk

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"mellium.im/sasl"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
)

type connectionErrHandler func(err error)

/*
Connect dials the server and starts receiving the events.
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

	if self.Session == nil {
		panic("session never got set")
	}

	if blocking {
		return self.startServing()
	} else {
		//serve in a thread
		go func() {
			err := self.startServing()

			//if error, try callback error handler, otherwise panic
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

func (self *XmppClient) SendText(to jid.JID, body string) error {
	msg := XMPPChatMessage{
		Message: stanza.Message{
			To:   to,
			Type: stanza.ChatMessage,
		},
		ChatMessageBody: ChatMessageBody{
			Body: &body,
		},
	}
	err := self.Session.Encode(self.Ctx, msg)
	return err
}

// CreateClient creates the client object using the login info object and returns it
func CreateClient(login *LoginInfo, dmHandler ChatMessageHandler) (*XmppClient, error) {
	// create client object
	client := &XmppClient{
		Login:     login,
		dmHandler: dmHandler,
	}
	client.Ctx, client.CtxCancel = context.WithCancel(context.Background())

	//client.MucClient
	messageNS := xml.Name{
		//Space: "jabber:client",
		Local: "body",
	}

	client.Multiplexer = mux.New(
		"jabber:client",
		muc.HandleClient(client.MucClient),
		mux.MessageFunc(stanza.ChatMessage, messageNS, client.internalHandleDM),
	)

	//string to jid object
	j, err := jid.Parse(login.User)
	if err != nil {
		return nil,
			errors.New("Could not parse user JID from `" + login.User + " - " + err.Error())
	}
	server := j.Domainpart()
	client.JID = &j
	client.Server = &server

	return client, nil
}
