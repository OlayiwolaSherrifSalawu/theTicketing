package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type contexK string

// tokenLifetime is how long a login lasts; the access cookie's MaxAge
// matches it so the browser drops the cookie when the token expires.
const tokenLifetime = 12 * time.Hour

func OpenDb(dsn *pq.Connector) (*sql.DB, error) {
	db := sql.OpenDB(dsn)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("%w: %v", OPENING_POOL_FAILED, err)
	}
	return db, nil
}

// MigrationsSource is where RunDBMigration reads migrations from,
// relative to the working directory. Tests in other directories override it.
var MigrationsSource = "file://database"

// RunDBMigration applies every pending migration in MigrationsSource.
func RunDBMigration(db *sql.DB) error {
	dbd, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	migra, err := migrate.NewWithDatabaseInstance(MigrationsSource, "postgres", dbd)
	if err != nil {
		return err
	}
	if err := migra.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("%w: %v", MIGRATION_FAILED, err)
	}
	return nil
}

func hashPassword(s string) (string, error) {
	hashByte, err := bcrypt.GenerateFromPassword([]byte(s), 10)
	if err != nil {
		return "", err
	}
	return string(hashByte), nil
}

func (a *Application) generateToken(id string) (string, error) {
	claims := jwt.MapClaims{
		"userID": id,
		"exp":    time.Now().Add(tokenLifetime).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.JWTSecret)
}

// getToken returns the raw JWT from the Authorization header (API
// clients) or, failing that, the access cookie (browsers).
func (a *Application) getToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := r.Cookie("access"); err == nil {
		return cookie.Value
	}
	return ""
}

func extractUserFromtoken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", FAILED_TO_GETID
	}
	userID, ok := claims["userID"].(string)
	if !ok || userID == "" {
		return "", FAILED_TO_GETID
	}
	return userID, nil
}

func extractUserFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ct).(string)
	return val, ok && val != ""
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// decodeJSON reads a JSON body of at most maxBytes into dst.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	return json.NewDecoder(r.Body).Decode(dst)
}
