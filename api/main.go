package api

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type AppInterface interface {
	Run()
}

type Application struct {
	Port        string
	ErroMessage *log.Logger
	InfoLogger  *log.Logger
	DB          *sql.DB
}

func NewApplication(db *sql.DB) *Application {
	return &Application{
		DB: db,
	}
}
func NewAppInterface(s *Application) AppInterface {
	return &Application{}
}

func CreateDataBase(){
	
}
func (a *Application) Run() {
	a.InfoLogger = log.New(os.Stdout, "INFO:  \t", log.Ldate|log.Ltime)
	a.ErroMessage = log.New(os.Stderr, "ERROR: \t", log.Ldate|log.Ltime)
	flag.StringVar(&a.Port, "port", ":4000", "Port Address")
	flag.Parse()
	// creating a server
	newMux := a.routers()
	serve := http.Server{
		Addr:     a.Port,
		Handler:  newMux,
		ErrorLog: a.ErroMessage,
	}
	err := serve.ListenAndServe()
	a.InfoLogger.Printf("started server on port %s\n", a.Port)
	a.ErroMessage.Println(err)
}

func OpenDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, OPENING_POOL_FAILED
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
