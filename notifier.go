package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxln "github.com/jackc/pgxlisten"
)

const (
	notificationChannel     = "job-channel"
	notificationChannelSize = 10
)

// Notifier LISTENs for notifications from PostgreSQL about newly added jobs,
// and distributes them to subscribers.
type Notifier struct {
	subs []chan struct{}
	ln   pgxln.Listener
}

func NewNotifier(pool *pgxpool.Pool) *Notifier {
	ln := pgxln.Listener{
		Connect: func(ctx context.Context) (*pgx.Conn, error) {
			conn, err := pool.Acquire(ctx)
			return conn.Conn(), err
		},
		LogError:       func(_ context.Context, err error) { log.Printf("ERROR: %v\n", err) },
		ReconnectDelay: 5 * time.Second,
	}
	return &Notifier{ln: ln}
}

func (n *Notifier) Run(ctx context.Context) error {
	defer n.stop()

	n.ln.Handle(
		notificationChannel,
		pgxln.HandlerFunc(func(ctx context.Context, notif *pgconn.Notification, _ *pgx.Conn) error {
			n.handle(ctx, notif)
			return nil
		}))
	return n.ln.Listen(ctx)
}

func (n *Notifier) Subscribe(_ context.Context) <-chan struct{} {
	ch := make(chan struct{}, notificationChannelSize)
	n.subs = append(n.subs, ch)
	return ch
}

func (n *Notifier) handle(ctx context.Context, notif *pgconn.Notification) {
	for _, subCh := range n.subs {
		subCh <- struct{}{}
	}
}

func (n *Notifier) stop() {
	for _, sub := range n.subs {
		close(sub)
	}
}
