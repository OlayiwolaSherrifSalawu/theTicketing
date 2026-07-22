package api

import (
	"log"
	"os"
)

func Run() {
	conn, errr := CreateConnection("config.env")
	if errr != nil {
		return
	}
	newLog := log.New(os.Stderr, "error-message", log.Ltime)
	sqlVAL, errr := OpenDb(conn)
	if errr != nil {
		newLog.Println(errr)
		return
	}
	app, err := NewApplication(sqlVAL)
	if err != nil {
		newLog.Println(err)
	}
	errors := app.RunDBMigration(sqlVAL)
	if errors != nil {
		newLog.Println(errors)
		return
	}
	apps := NewAppInterface(app)

	apps.Run()
}
