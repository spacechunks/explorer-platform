package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

/*
{
  "type": "PUSH_ARTIFACT",
  "occur_at": 1701565234,
  "operator": "test",
  "event_data": {
    "resources": [
      {
        "digest": "sha256:189d0d24eedfe5a03d867f47a1a75a552d82c6a479a84903520706b3ac6c3b65",
        "tag": "1337",
        "resource_url": "reg1.chunks.76k.io/proggers/pg:1337"
      }
    ],
    "repository": {
      "date_created": 1701495306,
      "name": "pg",
      "namespace": "proggers",
      "repo_full_name": "proggers/pg",
      "repo_type": "private"
    }
  }
}
*/

// skip writing tests since this a tmp solution

type Payload struct {
	OccurAt   uint64 `json:"occur_at"`
	Operator  string `json:"operator"`
	EventData struct {
		Resources []Resource `json:"resources"`
	} `json:"event_data"`
	Repository struct {
		RepoFullName string `json:"repo_full_name"`
	} `json:"repository"`
}

type Resource struct {
	Digest      string `json:"digest"`
	Tag         string `json:"tag"`
	ResourceURL string `json:"resource_url"`
}

type Handler func(p Payload)

func ListenHTTP(addr string, h Handler) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handle(h))
	return http.ListenAndServe(addr, mux)
}

func handle(h Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("read body: %v", err)
			return
		}
		var p Payload
		if err := json.Unmarshal(b, &p); err != nil {
			log.Printf("marshal json: %v", err)
			return
		}
		//fmt.Println(string(b))
		h(p)
	}
}
