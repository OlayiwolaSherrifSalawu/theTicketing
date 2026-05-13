package main

import (
	"database/sql"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/api"
)

func main() {
	var db *sql.DB
	app := api.NewApplication(db)
	apps := api.NewAppInterface(app)
	apps.Run()
}
