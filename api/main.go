package api

import (
	"log"
	"os"
)

func Run() {
	errLog := log.New(os.Stderr, "ERROR: \t", log.Ldate|log.Ltime)

	conn, err := CreateConnection("config.env")
	if err != nil {
		errLog.Fatal(err)
	}
	db, err := OpenDb(conn)
	if err != nil {
		errLog.Fatal(err)
	}
	defer db.Close()

	if err := RunDBMigration(db); err != nil {
		errLog.Fatal(err)
	}
	app, err := NewApplication(db, os.Getenv("JWT"))
	if err != nil {
		errLog.Fatal(err)
	}

	NewAppInterface(app).Run()
}
