package chunk

/*
import (
	"fmt"
	"github.com/spacechunks/platform/internal/image"
	"log"
)

func ProcessImage(src OCISource, dst OCISource, configPath string) (Config, error) {
	log.Printf("pulling img=%s\n", src.Ref())
	img, err := image.Pull(src.Ref(), src.User, src.Pass)
	if err != nil {
		return Config{}, fmt.Errorf("pull image: %w", err)
	}

	confFile, err := image.UnpackFile(src, configPath)
	if err != nil {
		return nil, fmt.Errorf("find config: %w", err)
	}
	var conf chunk.Config
	if err := yaml.Unmarshal(confFile.Content, &conf); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	flavorImgs, err := image.RepackFlavors(img, configPath)
	if err != nil {
		return Config{}, fmt.Errorf("repack: %w", err)
	}

	// reg1.chunks.76k.io/internalnamespace/freggy/bedwars
	dst.Repo = fmt.Sprintf("%s/%s", dst.Repo, src.Repo)
	// v1.0.0-2x2
	dst.Tag = fmt.Sprintf("%s-%s", src.Tag, v.ID)
	log.Printf("pushing variant img=%s\n", dst.Ref())
	if err := image.Push(varImg, dst.Ref(), dst.User, dst.Pass); err != nil {
		return Config{}, fmt.Errorf("img push: %w", err)
	}

	return conf, nil
}*/
