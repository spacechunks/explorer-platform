package chunk

import (
	"fmt"
	"github.com/chunks76k/internal/image"
	"gopkg.in/yaml.v3"
	"log"
)

func ProcessImage(imgRef, configPath, user, pass string) error {
	log.Printf("pulling img=%s\n", imgRef)
	img, err := image.Pull(imgRef, user, pass)
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	f, err := image.ExtractFile(img, configPath)
	if err != nil {
		return fmt.Errorf("find config: %w", err)
	}
	var conf Config
	if err := yaml.Unmarshal(f.Content, &conf); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	// for each variant create a new image
	// using the pushed one as its base
	for _, v := range conf.Variants { // TODO: use go routines (240KMH!!!!)
		log.Printf("variant=%s", v.Name)
		files, err := image.ExtractDir(img, v.Path)
		if err != nil {
			return fmt.Errorf("extract dir: %w", err)
		}
		for i := range files {
			f := &files[i]
			// adjust absolute path so variant files end up
			// in the server root directory when appending the layer.
			f.AbsPath = fmt.Sprintf("%s%s", conf.ServerRoot, f.RelPath)
		}
		varImg, err := image.AppendLayerFromFiles(img, files)
		if err != nil {
			return fmt.Errorf("append layer: %w", err)
		}
		// reg1.chunks.76k.io/freggy/bedwars:v1.20-2x2
		varRef := fmt.Sprintf("%s-%s", imgRef, v.Name)
		log.Printf("pushing variant img=%s\n", varRef)
		if err := image.Push(varImg, varRef, user, pass); err != nil {
			return fmt.Errorf("img push: %w", err)
		}
	}
	return nil
}
