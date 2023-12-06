package chunk

type DBConfig struct {
	User string
	Pass string
	URL  string
}

type Config struct {
	Version    string    `yaml:"version"`
	Name       string    `yaml:"name"`
	ServerRoot string    `yaml:"serverRoot"`
	Variants   []Variant `yaml:"variants"`
}

type Variant struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	Replicas int    `yaml:"replicas"`
}
