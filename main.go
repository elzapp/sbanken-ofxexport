package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/elzapp/go-ofx"

	"github.com/elzapp/go-sbanken"
)

func main() {
	config, err := ioutil.ReadFile("config.json")

	var creds sbanken.Credentials

	if err != nil {
		fmt.Println("You need a config.json, here's an example config.json:")
		j, _ := json.MarshalIndent(&creds, "", "  ")
		fmt.Println(string(j))
		os.Exit(1)
	}
	json.Unmarshal(config, &creds)
	client := sbanken.NewAPIConnection(creds)
	for _, account := range client.GetAccounts() {
		var ofxlist ofx.OfxTransactionList
		ofxlist.CurDef = "NOK"
		ofxlist.PayerAccount = account.AccountNumber
		ofxlist.PayerBank = "Sbanken"
		f, _ := os.Create(account.AccountNumber + ".ofx")
		defer f.Close()

		nonArchive := 0
		archive := 0
		for _, tx := range client.GetTransactions(account.AccountID) {
			if tx.Source == "Archive" {
				var btx ofx.BankTransaction
				btx.Amount = tx.Amount
				btx.DestinationAccount = tx.OtherAccountNumber
				btx.InterestDate = tx.GetInterestDate()
				btx.PostedDate = tx.GetAccountingDate()
				btx.Memo = tx.Text
				btx.Ref = tx.TransactionID
				ofxlist.Transactions = append(ofxlist.Transactions, btx.ToOfx())
				archive++
			} else {
				nonArchive++
			}
		}
		fmt.Printf("Saved %d transactions from %s\n", archive, account.AccountNumber)
		fmt.Printf("%s have %d transactions pending\n", account.AccountNumber, nonArchive)

		ofxlist.WriteOFX(f)
	}

}
