package api

func Run() {
	conn, errr := CreateConnection("config.env")
	if errr != nil {
		return
	}
	sqlVAL, errr := OpenDb(conn)
	if errr != nil {
		return
	}
	app := NewApplication(sqlVAL)

	apps := NewAppInterface(app)
	apps.Run()
}
