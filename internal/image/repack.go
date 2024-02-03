package image

import (
	"fmt"
	ociv1 "github.com/google/go-containerregistry/pkg/v1"
	"log"
)

func Repack(src ociv1.Image, rootPath string, paths []string) ([]ociv1.Image, error) {
	imgs := make([]ociv1.Image, 0, len(paths))
	// for each variant create a new image
	// using the pushed one as its base
	for _, path := range paths { // TODO: use go routines (240KMH!!!!)
		log.Printf("extract flavor=%s\n", path)
		files, err := UnpackDir(src, path)
		log.Printf("extracted flavor=%s\n", path)
		if err != nil {
			return nil, fmt.Errorf("extract dir: %w", err)
		}
		for i := range files {
			// adjust absolute path so flavor files end up
			// in the server root directory when appending the layer.
			files[i].AbsPath = fmt.Sprintf("%s%s", rootPath, files[i].RelPath)
		}
		log.Printf("append flavor=%s\n", path)
		varImg, err := AppendLayerFromFiles(src, files)
		log.Printf("appended flavor=%s\n", path)
		if err != nil {
			return nil, fmt.Errorf("append layer: %w", err)
		}
		imgs = append(imgs, varImg)
	}
	return imgs, nil
}
