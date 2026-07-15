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

func (d *TheTicketModel) GetByEmail(ctx context.Context, Email string, User *model.User) error {
	query := "select id, username, email_address, created_at, password from users where e	mail_address = $1"
	err := d.DB.QueryRowContext(ctx, query, Email).Scan(&User.ID, &User.UserName, &User.EmailAddress, &User.TimeStamp, &User.HashPassword)
	if err != nil {
		return err
	}
	return nil
}
