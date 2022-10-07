package db

import (
	"github.com/go-sql-driver/mysql"
)

func IsDup(err error) bool {
	if err == nil {
		return false
	}
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		if mysqlError.Number == 1062 { //ER_DUP_ENTRY https://mariadb.com/kb/en/mariadb-error-codes/
			return true
		} else {
			log.Infof("SQL Error: %d", mysqlError.Number)
		}
	}
	return false
}
