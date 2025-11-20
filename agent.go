package main

import (
	"context"
	"os"

	"github.com/567-labs/instructor-go/pkg/instructor"
	"github.com/567-labs/instructor-go/pkg/instructor/core"
	instructor_openai "github.com/567-labs/instructor-go/pkg/instructor/providers/openai"
	"github.com/sashabaranov/go-openai"
)
const sample_input = "The transaction on your IndusInd Bank Credit Card ending 0919 for INR 776.45 on 28-09-2025 09:01:28 pm at Eazydiner Private Limi is Approved. Available Limit: INR 77,947.00. \r\n\r\n In case you have not authorized this transaction, you can block your Credit Card instantly by sending SMS BLOCK 0919 to 5676757 from your registered mobile number. Please call 18602677777 for further queries.\r\n\r\n \r\n\r\n Take full control\r\n\r\n of your Credit Card with the all new IndusMobile App Download Now\r\n\r\n It's Offers Galore!\r\n\r\n Save big with these exciting offers on a wide range of brands View All Offers\r\n\r\n Redeem your reward points\r\n\r\n and avail offers on Spa and F\u0026B with  Visit Now\r\n\r\n          \r\n \r\nThis is an auto generated email, please do not reply to this email. IndusInd Bank Phone Banking services can be reached at 18602677777 from anywhere in the country through landline or mobile phone. Customers calling from outside India may call +91-22-42207777 through landline or mobile phone."


const extractor_prompt = `
- Role: You are an information extraction agent. You receive only the email body as a plain text string (no headers or subject).
- Task: From this body text, extract only the primary payment/transfer amount, its currency, and the recipient name.
- Rules:
	- Output exactly one JSON object with these three fields and nothing else:
	{
	
	"amount": number|null,
	
	"currency": string|null,
	
	"recipient": string|null
	
	}
	- amount: numeric value without commas; parse and normalize formats like “$1,299.00” -> 1299.0. If ambiguous or absent, use null.
	- currency: 3-letter ISO code if present (USD, EUR, INR, etc.); otherwise infer from symbols ($, €, ₹, £) or obvious locale indicators in the body text. Use null if unsure.
	- recipient: the person or entity receiving money (e.g., “Payee”, “Beneficiary”, “To”, “Credited to”). Preserve original casing; null if not determinable from the body.
	- If multiple amounts exist, choose the primary payment/transfer amount. Prefer the amount near cues like “total”, “paid”, “amount due”, “transfer”, “credited”, “sent”, “invoice total”. If still unclear, set amount and currency to null.
	- Do not infer from anything outside the provided body string. No commentary or extra fields—return only the JSON object.
	- If the body is not about a payment/transfer, return:
	{ "amount": null, "currency": null, "recipient": null }
`
type AgentResp struct {
	Amount		float64 	`json:"amount" jsonschema:"title=the amount,description=The amount spent in the transaction,example=100.00,example=1500.23"`
	Recipient  	string   	`json:"recipient" jsonschema:"title=the recipient,description=The recipient to whom the money was transfered or spent on,example=Netflix,example=VPA@shreya"`
	Currency	string		`json:"currency" jsonschema:"title=the currency,description=The currency of the amount spent in the transaction,example=Rs,example=INR,example=USD"`
}

func get_agent_response(client *instructor.InstructorOpenAI,emailBody string) (AgentResp,error) {
	// fmt.Printf("--------------Email Body------------\n%s",emailBody)
	conversation := core.NewConversation(extractor_prompt)
	
	conversation.AddUserMessage(emailBody)
	
	var agentresp AgentResp
	resp,err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4Dot1Nano,
			Messages: instructor_openai.ConversationToMessages(conversation),
		},
		&agentresp,
	)
	_ = resp
	
	if err != nil {
		return AgentResp{},err
	}
	// fmt.Printf("\n-------------Agent Response--------------\n%+v",agentresp)
	return agentresp,nil
}

func agent_init() (*instructor.InstructorOpenAI) {
	
	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)	
	
	return client
	// conversation.AddUserMessage(sample_input)
	
	// var amountrec AmountRec
	
	// resp, err := client.CreateChatCompletion(
	// 		ctx,
	// 		openai.ChatCompletionRequest{
	// 			Model:    openai.GPT4Dot1Nano,
	// 			Messages: instructor_openai.ConversationToMessages(conversation),
	// 		},
	// 		&amountrec,
	// 	)
	// 	_ = resp // sends back original response so no information loss from original API
	// 	if err != nil {
	// 	panic(err)
	// }
	
	// fmt.Printf("%+v\n", amountrec)
}

