package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

const mysqlDuplicateEntry = 1062

// IsDuplicateKeyError checks if err is a MySQL duplicate key error.
// Used to detect short code collisions and retry with a new code.
func IsDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == mysqlDuplicateEntry
	}
	return false
}
