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
type OriginID struct {
	XMLName xml.Name `xml:"urn:xmpp:sid:0 origin-id"`
	ID      string   `xml:"id,attr"`
}

type ChatMessageBody struct {
	Body     string    `xml:"body"`
	OriginID OriginID  `xml:"origin-id"`
	Reply    Reply     `xml:"reply"`
	Fallback Fallback  `xml:"fallback"`
	Request  *struct{} `xml:"request"`
	Markable *struct{} `xml:"markable"`
}

/*
XMPPChatMessage struct is a representation of the stanza such that it's contextual items
such as room, as well as abstract methods such as .reply()
*/
type XMPPChatMessage struct {
	Header   stanza.Message
	ChatBody ChatMessageBody
}

type ChatMessageHandler func(message XMPPChatMessage)

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
