package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/denpeshkov/pgqueue/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxln "github.com/jackc/pgxlisten"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	url, ok := os.LookupEnv("POSTGRESQL_URL")
	if !ok {
		log.Fatal("missing POSTGRESQL_URL env")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	ln := pgxln.Listener{
		Connect: func(ctx context.Context) (*pgx.Conn, error) {
			conn, err := pool.Acquire(ctx)
			return conn.Conn(), err
		},
		LogError:       func(_ context.Context, err error) { log.Printf("ERROR: %v\n", err) },
		ReconnectDelay: 5 * time.Second,
	}

	q := sqlc.New(pool)

	// TODO [pgxln.BacklogHandler]
	h := pgxln.HandlerFunc(func(ctx context.Context, notif *pgconn.Notification, conn *pgx.Conn) error {
		jobs, err := q.JobGetAvailable(ctx, 100)
		if err != nil {
			return err
		}
		for _, j := range jobs {
			log.Printf("%s [%s]\n", j.Description.String, j.State)
		}
		return nil
	})
	ln.Handle("channel", h)

	log.Fatal(ln.Listen(ctx))
}
