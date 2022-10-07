package db

import (
	"fmt"

	"github.com/go-msvc/errors"
)

type UserGroup struct {
	ID     string `json:"id" db:"id"`
	Title  string `json:"title" db:"title"`
	Status string `json:"status" db:"status"`
}

func ListUserGroups(userID string, filter string, limit int) ([]UserGroup, error) {
	sql := "SELECT g.id,g.title,gp.status" +
		" FROM persons as p" +
		" JOIN group_persons as gp ON gp.person_id=p.id" +
		" JOIN groups as g ON g.id=gp.group_id" +
		" WHERE p.auth_user_id=?"
	args := []interface{}{userID}
	if filter != "" {
		sql += " AND title like ?"
		args = append(args, "%"+filter+"%")
	}
	sql += fmt.Sprintf(" LIMIT %d", limit)
	var list []UserGroup
	if err := db.Select(&list, sql, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to get user groups")
	}
	return list, nil
}
