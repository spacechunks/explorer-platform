package main

import (
	"context"
	"github.com/chunks76k/internal/chunk"
	"github.com/chunks76k/internal/webhook"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"time"
)

const (
	configPath = "/app/.chunks.yaml"
	regHost    = "reg1.chunks.76k.io"
)

func main() {
	var (
		err error
		ctx = context.Background()
	)
	pool, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("unable to create connection pool: %v\n", err)
	}
	defer pool.Close()
	if err := webhook.ListenHTTP(":5080", func(p webhook.Payload) {
		for _, r := range p.EventData.Resources {
			// FIXME: more efficient retry logic
			for {
				if err = chunk.ProcessImage(r.ResourceURL, configPath, "test", "Test1234"); err != nil {
					log.Printf("process img: %v", err)
					continue
				}
				conn, err := pool.Acquire(ctx)
				if err != nil {
					log.Printf("cannot acquire db conn: %v", err)
					continue
				}
				if err = chunk.DeployAll(ctx, regHost, chunk.Config{}, conn.Conn()); err != nil {
					log.Printf("process img: %v", err)
					continue
				}
				if err == nil {
					return
				}
				// wait a second so things might have recovered
				time.Sleep(1 * time.Second)
			}
		}
	}); err != nil {
		log.Fatalf("listen http: %v", err)
	}
}
