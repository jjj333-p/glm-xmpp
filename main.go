package main

import (
	"encoding/json"
	"fmt"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"os"
	oasisSdk "pain.agency/oasis-sdk"
)

type glmConfig struct {
	LoginInfo oasisSdk.LoginInfo `json:"login_info"`
	//specific to this project
	llmInfo struct {
		Model        string `json:"LlmModel"`
		BaseURL      string `json:"LlmBaseURL"`
		ApiKey       string `json:"LlmApiKey"`
		SystemPrompt string `json:"SystemPrompt"`
	} `json:"llm_info"`
}

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llmResponse struct {
	Choices []struct {
		Message struct {
			Content string      `json:"content"`
			Refusal interface{} `json:"refusal"` // or *string if you expect a string/null
		} `json:"message"`
	} `json:"choices"`
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
		panic("Unable to read login.json - " + err.Error())
	}
	xmppConfig := glmConfig{}
	if err := json.Unmarshal(loginJSONbytes, &xmppConfig); err != nil {
		panic("Could not parse login.json - " + err.Error())
	}

	//sp := llmMessage{Role: "system", Content: xmppConfig.llmInfo.Model}
	//systemPrompt := []llmMessage{sp}
	//
	client, err := oasisSdk.CreateClient(&xmppConfig.LoginInfo, handleDM, handleGroupMessage, handleChatstate)
	if err != nil {
		panic("Could not create client - " + err.Error())
	}

	err = client.Connect(true, nil)
	if err != nil {
		panic("Could not connect - " + err.Error())
	}

}
