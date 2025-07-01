package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/denpeshkov/pgqueue/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
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

	n := NewNotifier(pool)

	notifCh := n.Subscribe(context.TODO())

	p := &Producer{db: sqlc.New(pool)}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return n.Run(ctx) })
	g.Go(func() error { return p.Run(ctx, notifCh) })

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
