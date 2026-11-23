package api

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/fs"
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
)

type AppInterface interface {
	Run()
}

type Application struct {
	Port          string
	ErroMessage   *log.Logger
	InfoLogger    *log.Logger
	DB            *sql.DB
	TheTicket     *theticket.TheTicketModel
	Templates     web.TemplateCache
	JWTSecret     []byte
	SecureCookies bool
}

// MinJWTSecretLen is the shortest signing key accepted for HS256. A short
// or empty key makes tokens trivially forgeable, so startup refuses it.
const MinJWTSecretLen = 32

// NewApplication wires the app together. jwtSecret must be at least
// MinJWTSecretLen bytes.
func NewApplication(db *sql.DB, jwtSecret string) (*Application, error) {
	if len(jwtSecret) < MinJWTSecretLen {
		return nil, fmt.Errorf("JWT secret must be at least %d bytes (set the JWT environment variable)", MinJWTSecretLen)
	}
	templateCache, err := web.NewTemplateCache()
	if err != nil {
		return nil, err
	}

	return &Application{
		DB:            db,
		TheTicket:     &theticket.TheTicketModel{DB: db},
		Templates:     templateCache,
		JWTSecret:     []byte(jwtSecret),
		SecureCookies: os.Getenv("CookieSecure") != "false",
		InfoLogger:    log.New(os.Stdout, "INFO:  \t", log.Ldate|log.Ltime),
		ErroMessage:   log.New(os.Stderr, "ERROR: \t", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}
func NewAppInterface(s *Application) AppInterface {
	return s
}

func (a *Application) Run() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
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

// CreateConnection builds a Postgres connector from environment variables
// (Host, Port, User, Password, Database). envFile is loaded first if it
// exists; a missing file is fine, so containers can pass plain env vars.
// Variables already set in the environment win over the file.
func CreateConnection(envFile string) (*pq.Connector, error) {
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("loading %s: %w", envFile, err)
	}
	port, _ := strconv.Atoi(os.Getenv("Port"))
	host := os.Getenv("Host")
	user := os.Getenv("User")
	password := os.Getenv("Password")
	database := os.Getenv("Database")
	// "prefer" uses TLS when the server offers it and falls back to plain
	// connections otherwise (as the local compose Postgres requires).
	sslMode := pq.SSLModePrefer
	if v := os.Getenv("SSLMode"); v != "" {
		sslMode = pq.SSLMode(v)
	}
	cfg := pq.Config{
		Host:           host,
		Port:           uint16(port),
		User:           user,
		ConnectTimeout: 5 * time.Second,
		Password:       password,
		Database:       database,
		SSLMode:        sslMode,
	}
	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}
