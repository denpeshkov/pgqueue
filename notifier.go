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
	chanSize            = 10
	notificationChannel = "job-channel"
)

type Subscription struct {
	Payload string
	closeCh chan<- struct{} // FIXME: handle errors
}

func (s Subscription) Done() {
	s.closeCh <- struct{}{}
}

type Notifier struct {
	subs []chan Subscription
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

func (n *Notifier) Subscribe(_ context.Context) <-chan Subscription {
	ch := make(chan Subscription, chanSize)
	n.subs = append(n.subs, ch)
	return ch
}

func (n *Notifier) handle(ctx context.Context, notif *pgconn.Notification) {
	for _, subCh := range n.subs {
		// FIXME:
		select {
		case <-ctx.Done():
			return
		default:
		}
		s := Subscription{
			Payload: notif.Payload,
			closeCh: make(chan<- struct{}),
		}
		subCh <- s
	}
}

func (n *Notifier) stop() {
	for _, sub := range n.subs {
		close(sub)
	}
}
