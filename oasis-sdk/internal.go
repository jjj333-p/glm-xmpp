package oasis_sdk

import (
	"encoding/xml"
	"mellium.im/xmlstream"
	"mellium.im/xmpp/stanza"
)

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

func (self *XmppClient) internalHandleDM(header stanza.Message, t xmlstream.TokenReadEncoder) error {
	//nothing to do if theres no handler
	if self.dmHandler == nil {
		return nil
	}

	//decode remaining parts to decode
	d := xml.NewTokenDecoder(t)
	body := &ChatMessageBody{}
	err := d.Decode(body)
	if err != nil {
		return err
	}
	msg := XMPPChatMessage{
		Header:   header,
		ChatBody: *body,
	}

	msg.ChatBody.ParseReply()

	//call handler and return to connection
	go self.dmHandler(msg)
	return nil
}
