package bank

import (
	"time"
)

type Transaction struct {
	Date    time.Time
	Amount  Amount
	Type    string
	Details string
	Code    string
}

func NewTransaction(date time.Time, amount Amount, txType string, details string, code string) Transaction {
	return Transaction{
		Date:    date,
		Amount:  amount,
		Type:    txType,
		Details: details,
		Code:    code,
	}
}
