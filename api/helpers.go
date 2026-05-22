package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type contexK string

func OpenDb(dsn *pq.Connector) (*sql.DB, error) {
	db := sql.OpenDB(dsn)
	if err := db.Ping(); err != nil {
		return nil, OPENING_POOL_FAILED
	}
	return db, nil
}

func (a *Application) RunDBMigration(db *sql.DB) {
	dbd, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		a.ErroMessage.Fatal(MIGRATION_FAILED, "Error:", err)
		return
	}
	migra, err := migrate.NewWithDatabaseInstance("file://database", "postgres", dbd)
	if err != nil {
		a.ErroMessage.Fatal(MIGRATION_FAILED, "Error:", err)
		return
	}
	err = migra.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			a.ErroMessage.Fatal(MIGRATION_FAILED, "Error:", err)
			return
		}
	}
}
func hashPassword(s string) (string, error) {
	hashByte, err := bcrypt.GenerateFromPassword([]byte(s), 10)
	if err != nil {
		return "", err
	}
	hashedPassword := string(hashByte)
	return hashedPassword, nil
}

func GenerateToken(id string) (string, error) {
	jwT := os.Getenv("JWT")
	claims := jwt.MapClaims{
		"userID": id,
		"exp":    time.Now().Add(time.Hour * 12).Unix(),
	}
	tokens := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tokens.SignedString([]byte(jwT))
	if err != nil {
		return "", err
	}

	return signed, err
}
func (a *Application) getToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		return strings.TrimPrefix(auth, "Bearer ")
	} else {
		cookie, err := r.Cookie("access")
		if err == nil {
			return cookie.Value
		}
	}

	return ""
}

func extractUserFromtoken(Token *jwt.Token) (string, error) {
	userId := ""
	if claims, ok := Token.Claims.(jwt.MapClaims); ok && Token.Valid {
		userId = claims["userID"].(string)
		return userId, nil
	}
	return "", FAILED_TO_GETID
}

func extractUserFromContext(ctx context.Context) (string, bool) {
	if val, ok := ctx.Value(ct).(string); ok {
		return val, true
	}
	return "", false
}
