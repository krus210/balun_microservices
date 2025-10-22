package repository

import (
	"errors"
	"fmt"
	"log"

	"users/internal/app/models"

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
			return fmt.Errorf("%s: %w", pgErr.Message, models.ErrAlreadyExists)
		default:
			return err
		}
	}
	return err
}
