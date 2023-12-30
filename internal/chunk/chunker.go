package chunk

import (
	"fmt"
	"github.com/chunks76k/internal/image"
	"gopkg.in/yaml.v3"
	"log"
)

func ProcessImage(src OCISource, dst OCISource, configPath string) (Config, error) {
	log.Printf("pulling img=%s\n", src.Ref())
	img, err := image.Pull(src.Ref(), src.User, src.Pass)
	if err != nil {
		return Config{}, fmt.Errorf("pull image: %w", err)
	}
	confFile, err := image.UnpackFile(img, configPath)
	if err != nil {
		return Config{}, fmt.Errorf("find config: %w", err)
	}
	var conf Config
	if err := yaml.Unmarshal(confFile.Content, &conf); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	// for each variant create a new image
	// using the pushed one as its base
	for _, v := range conf.Variants { // TODO: use go routines (240KMH!!!!)
		log.Printf("variant=%s", v.ID)
		log.Printf("extract img=%s\n", src.Ref())
		files, err := image.UnpackDir(img, v.Path)
		log.Printf("extracted img=%s\n", src.Ref())
		if err != nil {
			return Config{}, fmt.Errorf("extract dir: %w", err)
		}
		for _, f := range files {
			// adjust absolute path so variant files end up
			// in the server root directory when appending the layer.
			f.AbsPath = fmt.Sprintf("%s%s", conf.ServerRoot, f.RelPath)
		}
		log.Printf("append img=%s\n", src.Ref())
		varImg, err := image.AppendLayerFromFiles(img, files)
		log.Printf("appended img=%s\n", src.Ref())
		if err != nil {
			return Config{}, fmt.Errorf("append layer: %w", err)
		}
		// reg1.chunks.76k.io/internalnamespace/freggy/bedwars
		dst.Repo = fmt.Sprintf("%s/%s", dst.Repo, src.Repo)
		// v1.0.0-2x2
		dst.Tag = fmt.Sprintf("%s-%s", src.Tag, v.ID)
		log.Printf("pushing variant img=%s\n", dst.Ref())
		if err := image.Push(varImg, dst.Ref(), dst.User, dst.Pass); err != nil {
			return Config{}, fmt.Errorf("img push: %w", err)
		}
	}
	return conf, nil
}
