package main

import (
	"context"
	"fmt"

	"github.com/denpeshkov/pgqueue/sqlc"
)

const (
	jobsBatchSize = 5
)

// Producer fetches jobs from the PostgreSQL queue and dispatches them to workers.
// It receives completed job results from workers.
type Producer struct {
	db *sqlc.Queries
}

func (p *Producer) Run(ctx context.Context, notifCh <-chan struct{}) error {
	jobsCh := make(chan []sqlc.Job)
	go p.handleNotifications(ctx, notifCh, jobsCh)
	for {
		select {
		case <-ctx.Done():
			return nil
		case jobs := <-jobsCh:
			for i, job := range jobs {
				fmt.Println(i, job.State, job.Description)
			}
			//TODO: case result := <-jobResultCh:
		}
	}
}

func (p *Producer) handleNotifications(ctx context.Context, notifCh <-chan struct{}, jobsCh chan<- []sqlc.Job) {
	for range notifCh {
		// Each notification can signal an insertion of multiple jobs, possibly greater than the jobsBatchSize.
		for {
			jobs, err := p.db.GetJobs(ctx, jobsBatchSize)
			if err != nil {
				//FIXME:
			}
			if len(jobs) == 0 {
				break
			}
			jobsCh <- jobs
		}
	}
}
