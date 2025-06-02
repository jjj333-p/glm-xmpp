package oasis_sdk

import (
	"context"
	"encoding/xml"
	"mellium.im/xmpp"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
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

type FallbackBody struct {
	Start int `xml:"start,attr"`
	End   int `xml:"end,attr"`
}

type Fallback struct {
	XMLName xml.Name     `xml:"urn:xmpp:fallback:0 fallback"`
	For     string       `xml:"for,attr"`
	Body    FallbackBody `xml:"body"`
}

type Reply struct {
	XMLName xml.Name `xml:"urn:xmpp:reply:0 reply"`
	ID      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
}

// OriginID provided by XEP-0359: Unique and Stable Stanza IDs
type OriginID struct {
	XMLName xml.Name `xml:"urn:xmpp:sid:0 origin-id"`
	ID      string   `xml:"id,attr"`
}

// DeliveryReceiptRequest provided by XEP-0333: Displayed Markers
type DeliveryReceiptRequest struct {
	XMLName xml.Name `xml:"urn:xmpp:receipts request"`
}

// ReadReceiptRequest provided by XEP-0184: Message Delivery Receipts
type ReadReceiptRequest struct {
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 markable"`
}

// ----- begin Chatstates --------
type GoneChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates gone"`
}
type ActiveChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates active"`
}
type InactiveChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates inactive"`
}
type ComposingChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates composing"`
}
type PausedChatstate struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates paused"`
}

// ----- end Chatstates --------

type UnknownElement struct {
	XMLName xml.Name
	Content string     `xml:",innerxml"`
	Attrs   []xml.Attr `xml:",any,attr"`
}

type ChatMessageBody struct {
	Body               *string                 `xml:"body"`
	OriginID           *OriginID               `xml:"origin-id"`
	StanzaID           *stanza.ID              `xml:"stanza-id"`
	Reply              *Reply                  `xml:"reply"`
	Fallback           []Fallback              `xml:"fallback"`
	Request            *DeliveryReceiptRequest `xml:"request"`
	Markable           *ReadReceiptRequest     `xml:"markable"`
	Unknown            []UnknownElement        `xml:",any"`
	GoneChatState      *GoneChatstate          `xml:"gone"`
	ActiveChatState    *ActiveChatstate        `xml:"active"`
	InactiveChatState  *InactiveChatstate      `xml:"inactive"`
	ComposingChatState *ComposingChatstate     `xml:"composing"`
	PausedChatState    *PausedChatstate        `xml:"paused"`
	FallbacksParsed    bool
	CleanedBody        string
	ReplyFallbackText  string
}

func (self *ChatMessageBody) RequestingDeliveryReceipt() bool {
	return self.Request != nil
}

func (self *ChatMessageBody) RequestingReadReceipt() bool {
	return self.Markable != nil
}

type ChatState int

const (
	ChatStateNone ChatState = iota
	ChatStateActive
	ChatStateInactive
	ChatStateComposing
	ChatStatePaused
	ChatStateGone
)

func (self *ChatMessageBody) GetChatstate() ChatState {
	if self.ComposingChatState != nil {
		return ChatStateComposing
	}
	if self.PausedChatState != nil {
		return ChatStatePaused
	}
	if self.ActiveChatState != nil {
		return ChatStateActive
	}
	if self.InactiveChatState != nil {
		return ChatStateInactive
	}
	if self.GoneChatState != nil {
		return ChatStateGone
	}
	return ChatStateNone
}

/*
XMPPChatMessage struct is a representation of the stanza such that it's contextual items
such as room, as well as abstract methods such as .reply()
*/
type XMPPChatMessage struct {
	stanza.Message
	ChatMessageBody
}

type ChatMessageHandler func(client *XmppClient, message XMPPChatMessage)

// XmppClient is the end xmpp client object from which everything else works around
type XmppClient struct {
	Ctx         context.Context
	CtxCancel   context.CancelFunc
	Login       *LoginInfo
	JID         *jid.JID
	Server      *string
	Session     *xmpp.Session
	Multiplexer *mux.ServeMux
	MucClient   *muc.Client
	dmHandler   ChatMessageHandler
}
