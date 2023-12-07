package main

import (
	"context"
	"fmt"
	"github.com/chunks76k/internal/chunk"
	"github.com/chunks76k/internal/webhook"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"strings"
	"time"
)

const (
	configPath  = "/app/.chunks.yaml"
	variantRepo = "ghcr.io/freggy"
)

func main() {
	/*
		d, err := os.ReadFile("hack/TODO")
		if err != nil {
			log.Fatalf("NO FILE FOUND")
		}
		log.Println(string(d))
		time.Sleep(5000 * time.Second)*/
	var (
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
			// TODO: retry x times then stop and fail
			for {
				ref, err := name.ParseReference(r.ResourceURL)
				if err != nil {
					log.Printf("parse ref: %v", err)
					continue
				}
				src := chunk.OCIArtifact{
					User: "test",
					Pass: "Test1234",
					// reg1.chunks.76k.io/<>/<>
					Repo: ref.Context().RepositoryStr(),
					URL:  ref.Context().RegistryStr(),
					Tag:  ref.Identifier(),
				}
				internalRepo := chunk.OCIArtifact{
					User: "freggy",
					Pass: "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",
					URL:  "ghcr.io",
					Repo: "freggy/internal",
				}

				m := chunk.Meta{
					// replace here to ensure we can use the id as a valid name everywhere
					ChunkID:      strings.Replace(ref.Context().RepositoryStr(), "/", "-", -1),
					ChunkVersion: ref.Identifier(),
				}

				chunkConf, err := chunk.ProcessImage(src, internalRepo, configPath)
				if err != nil {
					log.Printf("process img: %v", err)
					continue
				}
				conn, err := pool.Acquire(ctx)
				if err != nil {
					log.Printf("cannot acquire db conn: %v", err)
					continue
				}
				baseRef := fmt.Sprintf("%s/%s", internalRepo.RepoURL(), src.Repo)
				if err = chunk.DeployAll(ctx, baseRef, m, chunkConf, conn.Conn()); err != nil {
					log.Printf("deploy all: %v", err)
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
