package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/jszwec/csvutil"
)

type Transaction struct {
	Id				string
	Date			string
	Amount			string
	Recipient		string
	Body			string
	InstrumentName	string
	BankName		string
	Last4			string
	MimeType		string
}

type CardRow struct {
	InstrumentName 	string `csv:"instrument_name"`
	BankName		string `csv:"bank_name"`
	Last4			string `csv:"last4"`
	FromEmail		string `csv:"from_email"`
	Subject			string `csv:"subject"`
	ExcludeWords	string `csv:"exclude_words"`
}

type State struct {
	CardConfig []CardRow
	Transactions []Transaction
}

func run_pipeline(){
	f,err := os.ReadFile("cards_config.csv")
	if err != nil {
		log.Fatalf("Unable to read the file %v",err)
	}
	
	var rows []CardRow
	err = csvutil.Unmarshal(f,&rows)
	if err != nil {
		log.Fatalf("Unable to unmarshal the file %v",err)
	}
	st := State{
		CardConfig: rows,
		Transactions: make([]Transaction, 0), // Initialize as empty slice
	}
	
	//auth
	srv,err := getService()
	if err != nil {
		log.Fatalf("Unable to get the service %v",err)
	}
	
	
	q , err := st.build_query("2025-09-01","2025-09-30")
	if err != nil {
		log.Fatal(err)
	}
	
	messages, err := st.fetch_emails(srv,q)
	if err != nil {
		log.Fatal(err)
	}
	
	err = st.parse_emails(messages)
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(st.Transactions)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Write JSON to file
	err = os.WriteFile("./test_data/transactions.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON to file: %v", err)
	}
}