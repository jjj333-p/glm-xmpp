package main

import (
	"encoding/json"
	"fmt"
	"mellium.im/xmpp/stanza"
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

type MessageBody struct {
	stanza.Message
	Body string `xml:"body"`
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

//existed when i had the xml decoding in a goroutine, didnt work because pointer deref
//type msgListener func(tokenReadEncoder xmlstream.TokenReadEncoder, start *xml.StartElement) error

func handleDM(client *oasisSdk.XmppClient, msg oasisSdk.XMPPChatMessage) {
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

	//oasis_sdk.CreateClient()
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	//temporary until db and login page exists
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

	//go func(msgChan chan oasisSdk.XMPPAbstractChatMessage) {
	//	for msg := range msgChan {
	//		fmt.Println(msg.Stanza.Body)
	//	}
	//}(client.CreateListener("", "", "", "", -1, false))

	if client.Connect(true, nil) != nil {
		panic("Could not connect - " + err.Error())
	}
	//
	//j, err := jid.Parse(xmppConfig.User)
	//if err != nil {
	//	panic("Could not parse user - " + err.Error())
	//}
	//
	//server := j.Domainpart()
	//
	//d := dial.Dialer{}
	//
	//conn, err := d.DialServer(ctx, "tcp", j, server)
	//if err != nil {
	//	panic("Could not connect stage 1 - " + err.Error())
	//}
	//
	//session, err := xmpp.NewSession(ctx, j.Domain(), j, conn, 0, xmpp.NewNegotiator(func(*xmpp.Session, *xmpp.StreamConfig) xmpp.StreamConfig {
	//	return xmpp.StreamConfig{
	//		Lang: "en",
	//		Features: []xmpp.StreamFeature{
	//			xmpp.BindResource(),
	//			xmpp.StartTLS(&tls.Config{
	//				ServerName: j.Domain().String(),
	//				MinVersion: tls.VersionTLS12,
	//			}),
	//			xmpp.SASL("", xmppConfig.Password, sasl.ScramSha1Plus, sasl.ScramSha1, sasl.Plain),
	//		},
	//		TeeIn:  nil,
	//		TeeOut: nil,
	//	}
	//}))
	//if err != nil {
	//	panic("Could not connect stage 2 - " + err.Error())
	//}
	//
	//// Send initial presence to let the server know we want to receive messages.
	//err = session.Send(ctx, stanza.Presence{Type: stanza.AvailablePresence}.Wrap(nil))
	//if err != nil {
	//	//return fmt.Errorf("Error sending initial presence: %w", err)
	//	panic("Error sending initial presence - " + err.Error())
	//}
	//
	////testMucClient := &muc.Client{}
	////m := mux.New("jabber:client", muc.HandleClient(testMucClient))
	//
	//go func() {
	//
	//	messageHistory := make(map[string][]llmMessage)
	//
	//	_ = session.Serve(
	//		xmpp.HandlerFunc(
	//			func(tokenReadEncoder xmlstream.TokenReadEncoder, start *xml.StartElement) error {
	//				decoder := xml.NewTokenDecoder(xmlstream.MultiReader(xmlstream.Token(*start), tokenReadEncoder))
	//
	//				if _, err := decoder.Token(); err != nil {
	//					return err
	//				}
	//
	//				// Ignore anything that's not a message. In a real system we'd want to at
	//				// least respond to IQs.
	//				switch start.Name.Local {
	//				case "message":
	//
	//					body := MessageBody{}
	//					err := decoder.DecodeElement(&body, start)
	//					if err != nil && err != io.EOF {
	//						fmt.Println("Error decoding element - " + err.Error())
	//						return nil
	//					}
	//
	//					// Don'tokenReadEncoder reflect messages unless they are chat messages and actually have a
	//					// body.
	//					// In a real world situation we'd probably want to respond to IQs, at least.
	//					if body.Body == "" {
	//						//stanza.
	//						fmt.Printf("empty fart %s from %s\n", body.Type, body.From.Bare().String())
	//						return nil
	//					}
	//					if body.Type != stanza.ChatMessage && body.Type != stanza.GroupChatMessage {
	//						//pass back message, creating new channel if not open
	//						fmt.Printf("pass fart %s: %s\n", body.From.Bare().String(), body.Body)
	//
	//						return nil
	//					}
	//					fmt.Printf("stanza fart %s %s: %s\n", body.Type, body.From.Bare().String(), body.Body)
	//
	//					//testMucJid, _ := jid.Parse("testing@group.pain.agency")
	//					resourcepart := body.From.Resourcepart()
	//					var role string
	//					if resourcepart == xmppConfig.DisplayName {
	//						role = "assistant"
	//					} else {
	//						role = "user"
	//					}
	//
	//					thisMsgForLLM := llmMessage{
	//						Role:    role,
	//						Content: resourcepart + ": " + body.Body,
	//					}
	//					mucHistory := messageHistory[body.From.Bare().String()]
	//					mucHistory = append(mucHistory, thisMsgForLLM)
	//					if len(mucHistory) > 50 {
	//						mucHistory = mucHistory[:50]
	//					}
	//					messageHistory[body.From.Bare().String()] = mucHistory
	//
	//					//dont respond to self
	//					if resourcepart == xmppConfig.DisplayName {
	//						return nil
	//					}
	//
	//					//check for and remove mention
	//					if strings.HasPrefix(body.Body, xmppConfig.DisplayName) {
	//						msgContent := strings.TrimPrefix(body.Body, xmppConfig.DisplayName)
	//						msgContent = strings.TrimPrefix(msgContent, ", ")
	//						msgContent = strings.TrimPrefix(msgContent, ": ")
	//
	//						if strings.ToLower(msgContent) == "forget" {
	//							messageHistory[body.From.Bare().String()] = nil
	//							go func() {
	//								reply := MessageBody{
	//									Message: stanza.Message{
	//										//ID:   body.Body,
	//										To:   body.From.Bare(),
	//										From: j,
	//										Type: body.Type,
	//									},
	//									Body: resourcepart + ", drinking to forget! 🍻",
	//								}
	//								//debug.Printf("Replying to message %q from %s with body %q", msg.ID, reply.To, reply.Body)
	//								sER := session.Encode(ctx, reply)
	//								//err = session.Send(ctx, tokenReadEncoder)
	//								if sER != nil {
	//									fmt.Println("Error farting element - " + sER.Error())
	//								}
	//							}()
	//							return nil
	//						}
	//
	//						//mgs := append(muc)
	//						data := map[string]interface{}{
	//							"model":    xmppConfig.Model,
	//							"messages": append(systemPrompt, mucHistory...),
	//							"stream":   false,
	//						}
	//
	//						bodyBytes, err := json.Marshal(data)
	//						fmt.Println(string(bodyBytes))
	//						if err != nil {
	//							go func(err2 error) {
	//								fmt.Println("Error marshalling JSON:", err)
	//								reply := MessageBody{
	//									Message: stanza.Message{
	//										//ID:   body.Body,
	//										To:   body.From.Bare(),
	//										From: j,
	//										Type: body.Type,
	//									},
	//									Body: err2.Error(),
	//								}
	//								//debug.Printf("Replying to message %q from %s with body %q", msg.ID, reply.To, reply.Body)
	//								sER := session.Encode(ctx, reply)
	//								//err = session.Send(ctx, tokenReadEncoder)
	//								if sER != nil {
	//									fmt.Println("Error farting element - " + sER.Error())
	//								}
	//								return
	//							}(err)
	//							return nil
	//						}
	//
	//						go func(thisBody string) {
	//
	//							quotePart := resourcepart + "\n"
	//							lenSoFar := 0
	//							for _, line := range strings.Split(thisBody, "\n") {
	//								if line == "" || line == "\r" {
	//									continue
	//								}
	//								quotePart += "> " + line + "\n"
	//								lenSoFar += len(line)
	//								if lenSoFar > 500 {
	//									break
	//								}
	//							}
	//
	//							// Create HTTP request
	//							req, err := http.NewRequest("POST", xmppConfig.BaseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	//							req.Header.Set("Authorization", "Bearer "+xmppConfig.ApiKey)
	//							req.Header.Set("Content-Type", "application/json")
	//
	//							// Send request
	//							client := &http.Client{}
	//							resp, err := client.Do(req)
	//							defer resp.Body.Close()
	//							respBody, err := io.ReadAll(resp.Body)
	//
	//							var res llmResponse
	//							err = json.Unmarshal(respBody, &res)
	//
	//							if err != nil {
	//								fmt.Println("Error creating request:", err)
	//								reply := MessageBody{
	//									Message: stanza.Message{
	//										//ID:   body.Body,
	//										To:   body.From.Bare(),
	//										From: j,
	//										Type: body.Type,
	//									},
	//									Body: quotePart + err.Error(),
	//								}
	//								//debug.Printf("Replying to message %q from %s with body %q", msg.ID, reply.To, reply.Body)
	//								err = session.Encode(ctx, reply)
	//								//err = session.Send(ctx, tokenReadEncoder)
	//								if err != nil {
	//									fmt.Println("Error farting element - " + err.Error())
	//								}
	//								return
	//							}
	//
	//							fmt.Println("Status Code:", resp.StatusCode)
	//							fmt.Println("Response:", string(respBody))
	//
	//							for _, r := range res.Choices {
	//								var botResponse string
	//								//interface can be any value, go has no truthyness like js
	//								if (r.Message.Refusal == false || r.Message.Refusal == nil) && r.Message.Content != "" {
	//									botResponse = r.Message.Content
	//								} else {
	//									botResponse = "{ The LLM refused to respond }"
	//								}
	//								//TODO: ID should be uuid
	//								//TODO: origin id
	//								reply := MessageBody{
	//									Message: stanza.Message{
	//										//ID:   body.Body,
	//										To:   body.From.Bare(),
	//										From: j,
	//										Type: body.Type,
	//									},
	//									Body: quotePart + botResponse,
	//								}
	//								//debug.Printf("Replying to message %q from %s with body %q", msg.ID, reply.To, reply.Body)
	//								err = session.Encode(ctx, reply)
	//								//err = session.Send(ctx, tokenReadEncoder)
	//								if err != nil {
	//									fmt.Println("Error farting element - " + err.Error())
	//								}
	//							}
	//
	//							if len(res.Choices) < 1 {
	//								reply := MessageBody{
	//									Message: stanza.Message{
	//										//ID:   body.Body,
	//										To:   body.From.Bare(),
	//										From: j,
	//										Type: body.Type,
	//									},
	//									Body: quotePart + "No response from the llm.\nResponse Code:" + strconv.Itoa(resp.StatusCode) + "\nResponse: " + string(respBody),
	//								}
	//								//debug.Printf("Replying to message %q from %s with body %q", msg.ID, reply.To, reply.Body)
	//								err = session.Encode(ctx, reply)
	//								//err = session.Send(ctx, tokenReadEncoder)
	//								if err != nil {
	//									fmt.Println("Error farting element - " + err.Error())
	//								}
	//							}
	//						}(body.Body)
	//					}
	//
	//					return nil
	//
	//				case "iq":
	//					body := stanza.IQ{}
	//					err := decoder.DecodeElement(&body, start)
	//					if err != nil && err != io.EOF {
	//						fmt.Println("Error decoding element - " + err.Error())
	//						return nil
	//					}
	//					fmt.Println(body)
	//				}
	//
	//				return nil
	//			},
	//		),
	//	)
	//}()
	//
	////time.Sleep(15 * time.Second)
	//fmt.Println("Joining Muc")
	//
	//testMucClient := &muc.Client{}
	//mux.New(
	//	"jabber:client",
	//	muc.HandleClient(testMucClient),
	//)
	//
	//for _, mjs := range xmppConfig.MucsToJoin {
	//	go func(mucJIDStr string) {
	//		mucJID, jidErr := jid.Parse(mucJIDStr + "/" + xmppConfig.DisplayName)
	//		if jidErr != nil {
	//			fmt.Println("Error parsing jid:", jidErr.Error())
	//		}
	//		_, mucErr := testMucClient.Join(ctx, mucJID, session)
	//		if mucErr != nil {
	//			fmt.Println("Error joining muc - " + mucErr.Error())
	//		}
	//	}(mjs)
	//}
	//
	////mJID, _ := jid.Parse("users@mellium.chat/" + xmppConfig.DisplayName)
	////_, mucErr := testMucClient.Join(ctx, mJID, session)
	////testMucJid, _ := jid.Parse("testing@group.pain.agency/" + xmppConfig.DisplayName)
	////_, mucErr = testMucClient.Join(ctx, testMucJid, session /*muc.Option(&muc.conf)*/)
	////if mucErr != nil {
	////	fmt.Println("Error joining muc - " + mucErr.Error())
	////}
	//
	////testMucClient.
	//
	////stop from exiting
	//s := sync.WaitGroup{}
	//s.Add(1)
	//s.Wait()
	//
	//return
}
