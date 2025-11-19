package main

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

func(st *State) build_query(start_date string, end_date string) (string,error) {
	q := ""
	
	for _ , r := range st.CardConfig {
		// Don't add empty transactions here - they will be added when emails are parsed
		excluded_words := ""
		subjects := strings.Split(r.Subject, "|")
		
		if(r.ExcludeWords != "") {
			excluded_words_arr := strings.Split(r.ExcludeWords,"|")
			excluded_words = fmt.Sprintf(" -{%s}",strings.Join(excluded_words_arr,","))
			// fmt.Println(r.ExcludeWords)
		}
		
		for _ , sub := range subjects {
			q += fmt.Sprintf("(from:%s subject:\"%s\"%s) OR ",r.FromEmail,sub,excluded_words)
		}
	}
	q = q[:len(q)-3]
	q = fmt.Sprintf("(after:%s before:%s) AND ",start_date,end_date) + q
	fmt.Println(q)
	return q,nil
}

func(st *State) fetch_emails(srv *gmail.Service,query string) ([]*gmail.Message,error) {
	response, err := srv.Users.Messages.List("me").Do(googleapi.QueryParameter("q",query))
	if err != nil {
		return nil,err
	}
	msgs := response.Messages
	// msgs = msgs[:5]
	
	messages := []*gmail.Message{}
	for _,m := range msgs {
		res,err := srv.Users.Messages.Get("me",m.Id).Do(googleapi.QueryParameter("format","full"))
		if err != nil {
			fmt.Printf("Error fetching messageid: %s",m.Id)
		}
		messages = append(messages,res)
	}
	// fmt.Print(response)
	return messages,nil
}

func (st *State) build_transaction(id, content, date, mimeType string) {
	for _, card := range st.CardConfig {
		if strings.Contains(content,card.Last4) {
			txn := Transaction {
				Id: id,
				Date: date,
				Amount: "",
				Recipient: "",
				Body: content,
				InstrumentName: card.InstrumentName,
				BankName: card.BankName,
				Last4: card.Last4,
				MimeType: mimeType,
			}
			
			st.Transactions = append(st.Transactions, txn)
			fmt.Printf("Transaction found for card ending in %s\n", card.Last4)
			fmt.Println("-----")
			fmt.Printf("%+v\n",txn)
		}
	}
}

func(st *State) parse_emails(messages []*gmail.Message) error {

	for _ , m := range messages {
		mimeType := m.Payload.MimeType
		var content string
		var err error
		
		switch mimeType {
		case "text/plain":
			content, err = parseTextPlain(m)
		case "text/html":
			content, err = parseTextHtml(m)
		case "multipart/alternative":
			content, err = parseMultipartAlternative(m)
		case "multipart/mixed":
			content, err = parseMultipartMixed(m)
		default:
			fmt.Printf("Unknown MIME type: %s\n", mimeType)
			continue
		}
		
		if err != nil {
			fmt.Printf("Error parsing message %s: %v\n", m.Id, err)
			continue
		}
		
		if content == "" {
			fmt.Printf("No content found in message %s\n", m.Id)
			continue
		}
		
		//remove links from content
		linkRegex := regexp.MustCompile(`(https?://\S+|www\.\S+)`)
		cleanContent := linkRegex.ReplaceAllString(content, "")
		cleanContent = strings.TrimSpace(cleanContent)
		// Get email date
		date := getHeaderValue(m.Payload.Headers, "Date")
		st.build_transaction(m.Id,cleanContent,date,mimeType)
	}
	
	fmt.Printf("Total transactions found: %d\n", len(st.Transactions))
	return nil
}