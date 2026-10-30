package data

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestSaveData(t *testing.T) {

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	err = conn.Ping(t.Context())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
}
