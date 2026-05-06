package api

import "database/sql"

type AppInterface interface {
	Run()
}

type Application struct{}

func NewAppInterface(s *Application) AppInterface {
	return &Application{}
}

func (a *Application) Run() {

}

func openDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgress", dsn)
	if err != nil {
		return nil, OPENING_POOL_FAILED
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
