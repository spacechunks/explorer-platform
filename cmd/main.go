package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/peterbourgon/ff"
	"github.com/spacechunks/chunks/internal/chunk"
	"github.com/spacechunks/chunks/internal/webhook"
	"log"
	"os"
	"strings"
	"time"
)

const (
	configPath = "/opt/paper/.chunks.yaml"
)

// TODO:
// NEXT STEPS
// * test create
// 	* inspect dns entry
// * test image tag update

func main() {
	var (
		_ = context.Background()
	)

	fs := flag.NewFlagSet("chunker", flag.ContinueOnError)
	var (
		//dbURL = fs.String("DB_URL", "", "database url")

		// source registry is where the fat base image lives
		srcRegUser = fs.String("SRC_OCI_REG_USER", "", "source oci registry user")
		srcRegPass = fs.String("SRC_OCI_REG_PASS", "", "source oci registry password")

		// destination registry is where the processed images live
		dstRegUser = fs.String("DST_OCI_REG_USER", "", "destination oci registry user")
		dstRegPass = fs.String("DST_OCI_REG_PASS", "", "destination oci registry password")
		dstRegURL  = fs.String("DST_OCI_REG_URL", "", "destination oci registry url")
	)

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("CHUNKER")); err != nil {
		log.Fatalf("ff parse: %v", err)
	}

	/*
		pool, err := pgxpool.New(ctx, *dbURL)
		if err != nil {
			log.Fatalf("unable to create connection pool: %v\n", err)
		}
		defer pool.Close()*/

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
				src := chunk.OCISource{
					User: *srcRegUser,
					Pass: *srcRegPass,
					/*User: "freggy",
					Pass: "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",*/
					// reg1.chunks.76k.io/ <my/repo> /myimage:version
					Repo: ref.Context().RepositoryStr(),
					// <reg1.chunks.76k.io> /my/repo/myimage:version
					URL: ref.Context().RegistryStr(),
					Tag: ref.Identifier(),
				}

				internalRepo := chunk.OCISource{
					User: *dstRegUser,
					Pass: *dstRegPass,
					URL:  *dstRegURL,
					Repo: "chunks-system",
					/*User: "freggy",
					Pass: "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",
					URL:  "ghcr.io",
					Repo: "freggy/internal",*/
				}
				m := chunk.Meta{
					// replace here to ensure we can use the id as a valid name everywhere
					ChunkID:      strings.Replace(ref.Context().RepositoryStr(), "/", "-", -1),
					ChunkVersion: ref.Identifier(),
				}
				conf, err := chunk.ProcessImage(src, internalRepo, configPath)
				if err != nil {
					log.Printf("process img: %v", err)
					continue
				}
				/*conn, err := pool.Acquire(ctx)
				if err != nil {
					log.Printf("cannot acquire db conn: %v", err)
					continue
				}*/
				baseRef := fmt.Sprintf("%s/%s", internalRepo.RepoURL(), src.Repo)
				if err = chunk.DeployAll(context.Background(), baseRef, m, conf); err != nil {
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
