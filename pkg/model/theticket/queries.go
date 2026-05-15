package theticket

import (
	"database/sql"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
)

type TheTicketModel struct {
	DB *sql.DB
}

func (d *TheTicketModel) Insert(User *model.User) error {
	stmt := "INSERT INTO users(id, username, email_address, created_at, password) VALUES ($1, $2, $3, $4, $5)"
	tx, err := d.DB.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(stmt, User.ID, User.UserName, User.EmailAddress, User.TimeStamp, User.HashPassword)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
