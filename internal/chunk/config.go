package chunk

import (
	"fmt"
)

type DBConfig struct {
	User string
	Pass string
	URL  string
}

type Config struct {
	// members below we read from the
	// users configuration
	Version    string    `yaml:"version"`
	ServerRoot string    `yaml:"serverRoot"`
	Variants   []Variant `yaml:"variants"`
}

type OCIArtifact struct { // TODO: better name
	User string
	Pass string
	URL  string
	Tag  string
	Repo string
}

func (o OCIArtifact) Ref() string {
	return fmt.Sprintf("%s/%s:%s", o.URL, o.Repo, o.Tag)
}

func (o OCIArtifact) RepoURL() string {
	return fmt.Sprintf("%s/%s", o.URL, o.Repo)
}

// Meta holds metadata about the users chunk.
// these fields will be populated by the backend
// and cannot be configured via config file.
type Meta struct {
	ChunkID      string
	ChunkVersion string
}

type Variant struct {
	ID       string `yaml:"id"`
	Path     string `yaml:"path"`
	Replicas int    `yaml:"replicas"`
}
