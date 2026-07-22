package api

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model/theticket"
	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/web"
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
	Templates   web.TemplateCache
}

func NewApplication(db *sql.DB) (*Application, error) {
	templateCache, err := web.NewTemplateCache()
	if err != nil {
		return nil, err
	}

	return &Application{
		DB:        db,
		TheTicket: &theticket.TheTicketModel{DB: db},
		Templates: templateCache,
	}, nil
}
func NewAppInterface(s *Application) AppInterface {
	return s
}

func (a *Application) Run() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
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
	go func() {
		a.InfoLogger.Printf("started server on port %s\n", a.Port)
		err := serve.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.ErroMessage.Println(err)
			return
		}
	}()
	<-quit
	a.InfoLogger.Println("Shutting down server..")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := serve.Shutdown(ctx)
	if err != nil {
		a.ErroMessage.Fatal(err)
		return
	}
	a.InfoLogger.Println("Server shutdown properly")
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
