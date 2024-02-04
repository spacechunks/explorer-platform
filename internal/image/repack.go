package image

import (
	"fmt"
	"log"

	ociv1 "github.com/google/go-containerregistry/pkg/v1"
)

func Repack(src ociv1.Image, rootPath string, path string) (ociv1.Image, error) {
	// for each variant create a new image
	// using the pushed one as its base
	log.Printf("extract flavor=%s\n", path)
	files, err := UnpackDir(src, path)
	log.Printf("extracted flavor=%s\n", path)
	if err != nil {
		return nil, fmt.Errorf("extract dir: %w", err)
	}
	for i := range files {
		// adjust absolute path so flavor files end up
		// in the server root directory when appending the layer.
		files[i].AbsPath = fmt.Sprintf("%s%s", rootPath, files[i].StrippedPath)
	}
	log.Printf("append flavor=%s\n", path)
	img, err := AppendLayerFromFiles(src, files)
	log.Printf("appended flavor=%s\n", path)
	if err != nil {
		return nil, fmt.Errorf("append layer: %w", err)
	}
	return img, nil
}
