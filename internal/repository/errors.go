package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound = errors.New("not found")

	ErrDuplicate = errors.New("duplicate entry")
)

const mysqlDuplicateEntry = 1062

func IsDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == mysqlDuplicateEntry
	}
	return false
}