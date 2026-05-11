package api

import (
	"database/sql"
	_"github.com/lib/pq"
)

type AppInterface interface {
	Run()
}

type Application struct {
	DB *sql.DB
}

func NewApplication(db *sql.DB) *Application {
	return &Application{
		DB: db,
	}
}
func NewAppInterface(s *Application) AppInterface {
	return &Application{}
}

func (a *Application) Run() {

}

func openDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, OPENING_POOL_FAILED
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
