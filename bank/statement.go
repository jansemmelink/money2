package bank

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/msf/logger"
	"github.com/google/uuid"
	"github.com/jansemmelink/money/db"
)

var log = logger.New("money").New("statement")

const (
	unknownExpenseAccountName = "Unknown expense"
	unknownIncomeAccountName  = "Unknown income"

	accountTypeExpense = "expense"
	accountTypeIncome  = "income"
)

type IStatement interface {
	WithBranchName(n string) IStatement
	WithBranchCode(c string) IStatement
	WithAccountNumber(n string) IStatement
	WithOpeningBalance(b Amount) IStatement
	WithClosingBalance(b Amount) IStatement
	WithTransaction(tx Transaction) IStatement
	BankName() string
	BranchName() string
	BranchCode() string
	AccountNumber() string
	OpenDate() time.Time
	OpenBalance() Amount
	CloseDate() time.Time
	CloseBalance() Amount
	Transactions() []Transaction

	Validate() error

	ImportToDb() (string, error) //return statements.id from db
}

func NewStatement(bankName string) IStatement {
	return &statement{
		bankName:     bankName,
		transactions: []Transaction{},
	}
}

type statement struct {
	databaseID     string //defined when imported into the db or loaded from the db
	bankName       string
	branchName     string
	branchCode     string
	accNumber      string
	openingBalance Amount
	closingBalance Amount
	transactions   []Transaction
}

type overlappingStatement struct {
	ID          string     `db:"id"`
	OpeningDate db.SqlTime `db:"opening_date"`
	ClosingDate db.SqlTime `db:"closing_date"`
}

func (s statement) WithBranchName(n string) IStatement {
	s.branchName = n
	return s
}

func (s statement) WithBranchCode(c string) IStatement {
	s.branchCode = c
	return s
}

func (s statement) WithAccountNumber(n string) IStatement {
	s.accNumber = n
	return s
}

func (s statement) WithOpeningBalance(b Amount) IStatement {
	s.openingBalance = b
	return s
}

func (s statement) WithClosingBalance(b Amount) IStatement {
	s.closingBalance = b
	return s
}

func (s statement) WithTransaction(tx Transaction) IStatement {
	s.transactions = append(s.transactions, tx)
	return s
}

func (s statement) Validate() error {
	//validate the transactions adds up to the difference between open/close balances
	total, _ := NewAmount(0)
	for _, tx := range s.transactions {
		total = total.Add(tx.Amount)
		log.Debugf("total %v -> %v", tx.Amount, total)
	}
	expectedSum := s.closingBalance.Sub(s.openingBalance)
	if total.mc != expectedSum.mc {
		return fmt.Errorf("open(%v) + tx.total(%v) = %v != close(%v) (diff=%v)",
			s.openingBalance,
			total,
			s.openingBalance.Add(total),
			s.closingBalance,
			s.closingBalance.Sub(s.openingBalance.Add(total)))
	}
	return nil
}

func (s statement) BankName() string      { return s.bankName }
func (s statement) BranchName() string    { return s.branchName }
func (s statement) BranchCode() string    { return s.branchCode }
func (s statement) AccountNumber() string { return s.accNumber }

func (s statement) OpenDate() time.Time {
	if len(s.transactions) == 0 {
		return time.Now()
	}
	return s.transactions[0].Date
}

func (s statement) OpenBalance() Amount { return s.openingBalance }

func (s statement) CloseDate() time.Time {
	if len(s.transactions) == 0 {
		return time.Now()
	}
	return s.transactions[len(s.transactions)-1].Date
}

func (s statement) CloseBalance() Amount        { return s.closingBalance }
func (s statement) Transactions() []Transaction { return s.transactions }

//after loading complete statement, call this to import it into the db
func (s statement) ImportToDb() (string, error) {
	if s.databaseID != "" {
		return "", errors.Errorf("statement(%s) already in db", s.databaseID)
	}
	if len(s.transactions) == 0 {
		return "", errors.Errorf("no transactions in statement")
	}

	//see if bank account already exists
	bankAccount, err := GetBankAccount(s.bankName, s.accNumber)
	if err != nil {
		return "", errors.Wrapf(err, "failed to look for bank account")
	}

	if bankAccount == nil {
		//not found, create new account and bank account
		account := Account{
			//ID:   uuid.New().String(),
			Name: s.bankName + ":" + s.accNumber,
			Type: "asset",
		}
		bankAccount = &BankAccount{
			//ID:            uuid.New().String(),
			AccountID:     account.ID,
			Account:       &account,
			BankName:      s.bankName,
			BranchName:    s.branchName,
			BranchCode:    s.branchCode,
			AccountNumber: s.accNumber,
		}
		if err := bankAccount.Save(); err != nil {
			return "", errors.Wrapf(err, "failed to save bank_account")
		}
		log.Infof("New bank account: %+v", bankAccount)
	} else {
		log.Infof("Existing bank account: %+v", bankAccount)
	}

	//get default unknown income/expence accounts to credit/debit
	//for all transactions in this statements
	unknownExpenseAccount, err := getOrCreateAccount(unknownExpenseAccountName, accountTypeExpense)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get default account")
	}
	unknownIncomeAccount, err := getOrCreateAccount(unknownIncomeAccountName, accountTypeIncome)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get default account")
	}

	//list existing statements overlapping this date range
	var ostList []overlappingStatement
	if err := db.Db().Select(&ostList, "SELECT `id`,`opening_date`,`closing_date` FROM `statements`"+
		" WHERE `bank_account_id`=?"+
		" AND `opening_date` >= ?"+
		" AND `closing_date` <= ?",
		bankAccount.ID,
		s.OpenDate(),
		s.CloseDate(),
	); err != nil && err != sql.ErrNoRows {
		return "", errors.Wrapf(err, "failed to list overlapping statements")
	}
	log.Infof("%d overlapping statements:", len(ostList))
	for _, ost := range ostList {
		log.Infof("  overlapping stm: %+v", ost)
	}

	//add transactions
	statementID := ""
	for _, tx := range s.transactions {
		//debit or credit the bank account
		//the other account is not yet known
		var dtAccountID string
		var ctAccountID string
		if tx.Amount.MilliCents() > 0 {
			//dt the bank account
			dtAccountID = bankAccount.Account.ID
			ctAccountID = unknownIncomeAccount.ID
		} else {
			//ct the bank account
			dtAccountID = unknownExpenseAccount.ID
			ctAccountID = bankAccount.Account.ID
		}

		//skip of date is covered by overlapping statement
		overlap := false
		for _, ost := range ostList {
			if !tx.Date.Before(time.Time(ost.OpeningDate)) && !tx.Date.After(time.Time(ost.ClosingDate)) {
				log.Infof("Date %s skipped, included in statement(%s)", tx.Date, ost.ID)
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}

		//create statement before importing the first transaction
		if statementID == "" {
			statementID = uuid.New().String()
			openingDate := s.transactions[0].Date
			closingDate := s.transactions[len(s.transactions)-1].Date
			sql := "INSERT INTO `statements` SET" +
				" id=?, bank_account_id=?, opening_date=?, opening_balance=?, closing_date=?, closing_balance=?"
			args := []interface{}{
				statementID,
				bankAccount.ID,
				openingDate,
				s.openingBalance,
				closingDate,
				s.closingBalance,
			}
			if result, err := db.Db().Exec(sql, args...); err != nil {
				return "", errors.Wrapf(err, "failed to insert statement record")
			} else {
				nrRows, _ := result.RowsAffected()
				if nrRows != 1 {
					return "", errors.Errorf("inserted %d instead of 1 row", nrRows)
				}
			}
		}

		transactionID := uuid.New().String()
		sql := "INSERT INTO `transactions` SET" +
			"  id=?, date=?, amount=?, dt_account_id=?, ct_account_id=?, statement_id=?, statement_type=?, statement_code=?, statement_details=?, notes=?"
		args := []interface{}{
			transactionID,
			tx.Date,
			tx.Amount,
			dtAccountID,
			ctAccountID,
			statementID,
			limitStringLen(tx.Type, 200),
			limitStringLen(tx.Details, 200),
			tx.Code,
			nil,
		}
		if result, err := db.Db().Exec(sql, args...); err != nil {
			// if err == sql.ErrDup {

			// }
			return "", errors.Wrapf(err, "failed to insert statement record")
		} else {
			nrRows, _ := result.RowsAffected()
			if nrRows != 1 {
				return "", errors.Errorf("inserted %d instead of 1 row", nrRows)
			}
		}
	}
	return statementID, nil
} //statement.ImportToDB()

func limitStringLen(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[0:maxLen]
	}
	return s
}

func getOrCreateAccount(name string, accountType string) (*Account, error) {
	acc, _ := GetAccountByName(name)
	if acc != nil {
		log.Infof("Existing %s account %+v", name, acc)
		return acc, nil
	}
	acc = &Account{
		Name: name,
		Type: accountType,
	}
	if err := acc.Save(); err != nil {
		return nil, errors.Wrapf(err, "failed to create account(%s)", name)
	}
	log.Infof("Created %s account %+v", name, acc)
	return acc, nil
} //getOrCreateAccount()
