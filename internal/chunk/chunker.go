package chunk

import (
	"fmt"
	"log"

	"github.com/spacechunks/platform/internal/image"
	"gopkg.in/yaml.v3"
)

func ProcessImagePush(src OCISource, dst OCISource, configPath string) (Config, error) {
	log.Printf("pulling img=%s\n", src.Ref())
	srcImg, err := image.Pull(src.Ref(), src.User, src.Pass)
	if err != nil {
		return Config{}, fmt.Errorf("pull image: %w", err)
	}
	confFile, err := image.UnpackFile(srcImg, configPath)
	if err != nil {
		return Config{}, fmt.Errorf("find config: %w", err)
	}
	var conf Config
	if err := yaml.Unmarshal(confFile.Content, &conf); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	// TODO: goroutines
	for _, f := range conf.Flavors {
		flavorImg, err := image.Repack(srcImg, conf.ServerRoot, f.Path)
		if err != nil {
			return Config{}, fmt.Errorf("repack: %w", err)
		}
		// reg1.chunks.76k.io/internalnamespace/freggy/bedwars
		dst.Repo = fmt.Sprintf("%s/%s", dst.Repo, src.Repo)
		// v1.0.0-2x2
		dst.Tag = fmt.Sprintf("%s-%s", src.Tag, f.ID)
		log.Printf("pushing variant img=%s\n", dst.Ref())
		if err := image.Push(flavorImg, dst.Ref(), dst.User, dst.Pass); err != nil {
			return Config{}, fmt.Errorf("img push: %w", err)
		}
	}
	return conf, nil
}
