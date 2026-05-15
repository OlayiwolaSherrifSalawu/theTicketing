package api

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model/theticket"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
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
	TheTicket   *theticket.TheTicketModel
}

func NewApplication(db *sql.DB) *Application {
	return &Application{
		DB:        db,
		TheTicket: &theticket.TheTicketModel{DB: db},
	}
}
func NewAppInterface(s *Application) AppInterface {
	return s
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
	a.InfoLogger.Printf("started server on port %s\n", a.Port)

	err := serve.ListenAndServe()
	a.ErroMessage.Println(err)

}

func CreateConnection(envFile string) (*pq.Connector, error) {
	errs := godotenv.Load(envFile)
	if errs != nil {
		return nil, errs
	}

	port, _ := strconv.Atoi(os.Getenv("Port"))
	host := os.Getenv("Host")
	user := os.Getenv("User")
	password := os.Getenv("Password")
	database := os.Getenv("Database")
	cfg := pq.Config{
		Host:           host,
		Port:           uint16(port),
		User:           user,
		ConnectTimeout: 5 * time.Second,
		Password:       password,
		Database:       database,
	}
	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}
