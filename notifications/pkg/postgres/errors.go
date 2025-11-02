package postgres

import (
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func ConvertPGError(err error) error {
	if err == nil {
		return nil
	}

	// https://github.com/jackc/pgx/wiki/Error-Handling
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Println(pgErr.Message) // => syntax error at end of input
		log.Println(pgErr.Code)    // => 42601

		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%s", pgErr.Message)
		default:
			return err
		}
	}
	return err
}
