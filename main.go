package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	oasisSdk "pain.agency/oasis-sdk"
)

type glmConfig struct {
	LoginInfo oasisSdk.LoginInfo `json:"login_info"`
}

func handleDM(client *oasisSdk.XmppClient, msg *oasisSdk.XMPPChatMessage) {
	var replyBody string
	if msg.ReplyFallbackText == nil {
		replyBody = "nil"
	} else {
		replyBody = *msg.ReplyFallbackText
	}
	err := client.MarkAsRead(msg)
	if err != nil {
		fmt.Printf("Error marking as read: %v\n", err)
	}
	err = client.ReplyToEvent(
		msg,
		fmt.Sprintf(
			"message \"%s\" replying to \"%s\"\n",
			*msg.CleanedBody,
			replyBody,
		),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func handleGroupMessage(client *oasisSdk.XmppClient, ch *muc.Channel, msg *oasisSdk.XMPPChatMessage) {
	var replyBody string
	if msg.ReplyFallbackText == nil {
		replyBody = "nil"
	} else {
		replyBody = *msg.ReplyFallbackText
	}
	err := client.MarkAsRead(msg)
	if err != nil {
		fmt.Printf("Error marking as read: %v\n", err)
	}
	err = client.ReplyToEvent(
		msg,
		fmt.Sprintf(
			"message \"%s\" replying to \"%s\"\n",
			*msg.CleanedBody,
			replyBody,
		),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func deliveryReceiptHandler(client *oasisSdk.XmppClient, from jid.JID, id string) {
	fmt.Printf("Delivered %s to %s\n", id, from.String())
}
func readReceiptHandler(client *oasisSdk.XmppClient, from jid.JID, id string) {
	fmt.Printf("%s has seen %s\n", from.String(), id)
}

func handleChatstate(_ *oasisSdk.XmppClient, from jid.JID, state oasisSdk.ChatState) {
	fromStr := from.String()
	switch state {
	case oasisSdk.ChatStateActive:
		fmt.Println(fromStr, "is active")
	case oasisSdk.ChatStateComposing:
		fmt.Println(fromStr, "is composing")
	case oasisSdk.ChatStatePaused:
		fmt.Println(fromStr, "has paused typing")
	case oasisSdk.ChatStateInactive:
		fmt.Println(fromStr, "is inactive")
	case oasisSdk.ChatStateGone:
		fmt.Println(fromStr, "has gone")
	default:
		fmt.Println(fromStr, "is in an unknown state.")
	}
}

func main() {

	loginJSONbytes, err := os.ReadFile("db/login.json")
	if err != nil {
		log.Fatalln("Unable to read login.json - " + err.Error())
	}
	xmppConfig := glmConfig{}
	if err := json.Unmarshal(loginJSONbytes, &xmppConfig); err != nil {
		log.Fatalln("Could not parse login.json - " + err.Error())
	}

	client, err := oasisSdk.CreateClient(
		&xmppConfig.LoginInfo,
		handleDM,
		handleGroupMessage,
		handleChatstate,
		deliveryReceiptHandler,
		readReceiptHandler,
	)
	if err != nil {
		log.Fatalln("Could not create client - " + err.Error())
	}

	err = client.Connect(true, nil)
	if err != nil {
		log.Fatalln("Could not connect - " + err.Error())
	}

}
