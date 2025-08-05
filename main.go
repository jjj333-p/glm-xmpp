package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

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

func handleGroupMessage(client *oasisSdk.XmppClient, CH *muc.Channel, msg *oasisSdk.XMPPChatMessage) {

	if CH == nil || CH.Me().Equal(msg.From) {
		return
	}

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
	err = client.SendText(
		msg.From.Bare(),
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

func deliveryReceiptHandler(_ *oasisSdk.XmppClient, from jid.JID, id string) {
	fmt.Printf("Delivered %s to %s\n", id, from.String())
}
func readReceiptHandler(_ *oasisSdk.XmppClient, from jid.JID, id string) {
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

	go func() {
		err = client.Connect()
		if err != nil {
			log.Fatalln("Could not connect - " + err.Error())
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter file path: ")
	filePath, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln("Error reading input:", err)
	}

	// Trim newline character from input
	filePath = filePath[:len(filePath)-1]

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	progress := make(chan oasisSdk.UploadProgress)

	go client.UploadFile(ctx, filePath, progress)
	for update := range progress {
		if update.Error != nil {
			log.Fatalln("Error uploading file - " + update.Error.Error())
		}
		if update.GetURL != "" {
			fmt.Println("file upload done, available at", update.GetURL)

			desc := "Uploaded image description"

			err := client.SendSingleFileMessage(
				jid.MustParse("testing@group.pain.agency"),
				update.GetURL,
				&desc,
			)

			fmt.Println(err)
			fmt.Println(desc)
		}
		fmt.Printf(
			"%d out of %d bytes uploaded, %.2f%% complete\n",
			update.BytesSent, update.TotalBytes, update.Percentage,
		)
	}

	l := sync.Mutex{}
	l.Lock()
	l.Lock()
}
