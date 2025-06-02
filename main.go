package main

import (
	"encoding/json"
	"fmt"
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
	err := client.ReplyToEvent(
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
	client, err := oasisSdk.CreateClient(&xmppConfig.LoginInfo, handleDM)
	if err != nil {
		panic("Could not create client - " + err.Error())
	}

	if client.Connect(true, nil) != nil {
		panic("Could not connect - " + err.Error())
	}

}
