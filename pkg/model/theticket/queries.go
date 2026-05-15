package theticket

import (
	"context"
	"database/sql"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

type TheTicketModel struct {
	DB *sql.DB
}

func (d *TheTicketModel) Insert(ctx context.Context, User *model.User) error {
	stmt := "INSERT INTO users(id, username, email_address, created_at, password) VALUES ($1, $2, $3, $4, $5)"
	tx, err := d.DB.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, stmt, User.ID, User.UserName, User.EmailAddress, User.TimeStamp, User.HashPassword)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
