package bank

import (
	"database/sql"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/jansemmelink/money/db"
)

type Account struct {
	ID   string
	Name string
	Type string
}

func GetAccounts(nameFilter string, typeFilter string, limit int) ([]Account, error) {
	sql := "SELECT * FROM accounts"
	args := []interface{}{}

	filters := []string{}
	if nameFilter != "" {
		filters = append(filters, "name like ?")
		args = append(args, "%"+nameFilter+"%")
	}
	if typeFilter != "" {
		filters = append(filters, "type like ?")
		args = append(args, "%"+typeFilter+"%")
	}
	if len(filters) > 0 {
		sql += " WHERE " + strings.Join(filters, " AND ")
	}

	sql += " ORDER BY name"
	if limit <= 0 {
		limit = 10
	}
	sql += " LIMIT ?"
	args = append(args, limit)
	var accList []Account
	if err := db.Db().Select(&accList, sql, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to get list of accounts(%s,%s,%d)", nameFilter, typeFilter, limit)
	}
	return accList, nil
}

func GetAccount(id string) (*Account, error) {
	var acc Account
	if err := db.Db().Get(&acc,
		"SELECT id,name,type FROM accounts WHERE id=?",
		id,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to select account")
	}
	return &acc, nil
}

func GetAccountByName(name string) (*Account, error) {
	var acc Account
	if err := db.Db().Get(&acc,
		"SELECT id,name,type FROM accounts WHERE name=?",
		name,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to select account")
	}
	return &acc, nil
}

func (acc *Account) Save() error {
	if acc.Name == "" {
		return errors.Errorf("missing name")
	}
	if acc.Type == "" {
		return errors.Errorf("missing type")
	}
	if acc.ID == "" {
		id := uuid.New().String()
		if _, err := db.Db().Exec("INSERT INTO `accounts` SET id=?, name=?, type=?",
			id,
			acc.Name,
			acc.Type,
		); err != nil {
			return errors.Wrapf(err, "failed to insert account")
		}
		acc.ID = id
		log.Infof("Inserted bank_account(%s)", acc.ID)
	} else {
		if result, err := db.Db().Exec("UPDATE `accounts` SET name=?,type=? WHERE id=?",
			acc.Name,
			acc.Type,
			acc.ID,
		); err != nil {
			return errors.Wrapf(err, "failed to update account")
		} else {
			if nr, _ := result.RowsAffected(); nr != 1 {
				return errors.Errorf("updated %d account rows")
			}
		}
	}
	return nil
} //Account.Save()
