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
	"strings"
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

func (self *XmppClient) ReplyToEvent(originalMsg XMPPChatMessage, body string) error {
	//pull out JIDs as per https://xmpp.org/extensions/xep-0461.html#usecases
	replyTo := originalMsg.From
	to := replyTo.Bare()

	//name to include in fallback
	var readableReplyTo string
	if originalMsg.Type == stanza.ChatMessage {
		readableReplyTo = to.String()
	} else if originalMsg.Type == stanza.GroupChatMessage {
		readableReplyTo = replyTo.Resourcepart()
	}

	timeAgo := "TODO ago"

	originalBody := *originalMsg.CleanedBody
	quoteOriginalBody := readableReplyTo + " | " + timeAgo + "\n> " + strings.ReplaceAll(originalBody, "\n", "\n> ") + "\n"

	//ID to use in reply as per https://xmpp.org/extensions/xep-0461.html#business-id
	var replyToID string
	if originalMsg.Type == stanza.GroupChatMessage {
		//TODO check if room advertizes unique ids, if not cannot reply in groupchat even if id is present
		if originalMsg.StanzaID == nil || originalMsg.StanzaID.By.String() != to.String() {
			return self.SendText(to, quoteOriginalBody+body)
		}
		replyToID = originalMsg.StanzaID.ID
	} else if originalMsg.OriginID != nil {
		replyToID = originalMsg.OriginID.ID
	} else {
		replyToID = originalMsg.ID
	}

	// <reply> as per https://xmpp.org/extensions/xep-0461.html#usecases
	replyStanza := Reply{
		To: replyTo.String(),
		ID: replyToID,
	}

	// <fallback> as per https://xmpp.org/extensions/xep-0461.html#compat
	replyFallback := Fallback{
		For: "urn:xmpp:reply:0",
		Body: FallbackBody{
			Start: 0,
			End:   len(quoteOriginalBody) - 1,
		},
	}

	b := quoteOriginalBody + body
	msg := XMPPChatMessage{
		Message: stanza.Message{
			To:   to,
			Type: originalMsg.Type,
		},
		ChatMessageBody: ChatMessageBody{
			Body:     &b,
			Reply:    &replyStanza,
			Fallback: []Fallback{replyFallback},
		},
	}
	return self.Session.Encode(self.Ctx, msg)
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
