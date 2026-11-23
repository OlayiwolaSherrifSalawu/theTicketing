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
	_, err := d.DB.ExecContext(ctx, stmt, User.ID, User.UserName, User.EmailAddress, User.TimeStamp, User.HashPassword)
	return translate(err)
}

func (d *TheTicketModel) GetByEmail(ctx context.Context, Email string, User *model.User) error {
	query := "SELECT id, username, email_address, created_at, password FROM users WHERE lower(email_address) = lower($1)"
	err := d.DB.QueryRowContext(ctx, query, Email).Scan(&User.ID, &User.UserName, &User.EmailAddress, &User.TimeStamp, &User.HashPassword)
	return translate(err)
}

func (d *TheTicketModel) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := "SELECT id, username, email_address, created_at, password FROM users WHERE id = $1"
	u := &model.User{}
	err := d.DB.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.UserName, &u.EmailAddress, &u.TimeStamp, &u.HashPassword)
	if err != nil {
		return nil, translate(err)
	}
	return u, nil
}
