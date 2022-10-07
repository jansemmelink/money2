package bank

import (
	"database/sql"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/jansemmelink/money/db"
)

type BankAccount struct {
	ID            string   `db:"id"`
	AccountID     string   `db:"account_id"`
	Account       *Account `db:"-"`
	BankName      string   `db:"bank_name"`
	BranchName    string   `db:"branch_name"`
	BranchCode    string   `db:"branch_code"`
	AccountNumber string   `db:"account_number"`
}

func GetBankAccount(bankName string, accountNumber string) (*BankAccount, error) {
	var ba BankAccount
	if err := db.Db().Get(&ba,
		"SELECT id, account_id, bank_name, branch_name, branch_code, account_number FROM bank_accounts WHERE bank_name=? AND account_number=?",
		bankName,
		accountNumber,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //not found
		}
		return nil, errors.Wrapf(err, "failed to select bank_account")
	}
	if acc, err := GetAccount(ba.AccountID); err != nil {
		return nil, errors.Wrapf(err, "failed in GetAccount(%s)", ba.AccountID)
	} else {
		ba.Account = acc
	}
	return &ba, nil
}

func (ba *BankAccount) Save() error {
	if ba.BankName == "" {
		return errors.Errorf("missing bank_name")
	}
	if ba.AccountNumber == "" {
		return errors.Errorf("missing account_number")
	}
	if ba.AccountID == "" && ba.Account == nil {
		return errors.Errorf("missing account")
	}

	if ba.AccountID == "" {
		//account ID not yet set
		//make sure linked to an account
		if ba.Account == nil {
			return errors.Errorf("bank_account does not have an account")
		}

		//make sure the linked account is saved
		if ba.Account.ID == "" {
			if err := ba.Account.Save(); err != nil {
				return errors.Wrapf(err, "failed to save account")
			}
		}
		ba.AccountID = ba.Account.ID
	}

	if ba.ID == "" {
		id := uuid.New().String()
		if _, err := db.Db().Exec("INSERT INTO `bank_accounts` SET id=?, account_id=?, bank_name=?, branch_name=?, branch_code=?, account_number=?",
			id,
			ba.Account.ID,
			ba.BankName,
			ba.BranchName,
			ba.BranchCode,
			ba.AccountNumber,
		); err != nil {
			return errors.Wrapf(err, "failed to insert bank_account")
		}
		ba.ID = id
		log.Infof("Inserted bank_account(%s)", ba.ID)
	} else {
		if result, err := db.Db().Exec("UPDATE `bank_accounts` SET bank_name=?, branch_name=?, branch_code=?, account_number=? WHERE id=?",
			ba.BankName,
			ba.BranchName,
			ba.BranchCode,
			ba.AccountNumber,
			ba.ID,
		); err != nil {
			return errors.Wrapf(err, "failed to update bank_account")
		} else {
			if nr, _ := result.RowsAffected(); nr != 1 {
				return errors.Errorf("updated %d account rows")
			}
		}
		log.Infof("Updated bank_account(%s)", ba.ID)
	}
	return nil
} //BankAccount.Save()
