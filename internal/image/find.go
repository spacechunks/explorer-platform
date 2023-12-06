package image

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	ociv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"io"
	"log"
	"slices"
	"strings"
)

// TODO: tests

type File struct {
	AbsPath string
	RelPath string
	Content []byte
	Size    int64
	Dir     bool
}

// FIXME: refactor code a little bit

func ExtractFile(img ociv1.Image, path string) (File, error) {
	layers, err := img.Layers()
	if err != nil {
		return File{}, fmt.Errorf("image layers: %w", err)
	}

	// we will find our config file most likely
	// in the latest layers that have been added.
	// new layers are appended to the end of the slice,
	// so we have to reverse it.
	slices.Reverse(layers)

	for _, l := range layers {
		data, err := l.Uncompressed()
		if err != nil {
			return File{}, fmt.Errorf("layer uncompressed: %w", err)
		}
		defer data.Close()
		tr := tar.NewReader(data)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return File{}, fmt.Errorf("tar next: %w", err)
			}
			if hdr.Typeflag == tar.TypeDir {
				continue
			}
			// tar deals with relative paths, so make em absolute
			abs := fmt.Sprintf("/%s", hdr.Name)
			if abs != path {
				continue
			}
			var content = &bytes.Buffer{}
			if _, err := io.Copy(content, tr); err != nil {
				return File{}, fmt.Errorf("copy config bytes: %w", err)
			}
			return File{
				AbsPath: abs,
				RelPath: strings.TrimPrefix(abs, path),
				Content: content.Bytes(),
				Size:    hdr.Size,
			}, nil
		}
	}

	return File{}, errors.New("file not found")
}

func ExtractDir(img ociv1.Image, path string) ([]File, error) {
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("image layers: %w", err)
	}

	// same reason as in ExtractFile
	slices.Reverse(layers)
	files := make([]File, 0)

	for _, l := range layers {
		data, err := l.Uncompressed() // TODO: do in goroutine
		if err != nil {
			return nil, fmt.Errorf("layer uncompressed: %w", err)
		}
		defer data.Close()
		tr := tar.NewReader(data)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("tar next: %w", err)
			}
			// tar deals with relative paths, so make em absolute
			abs := fmt.Sprintf("/%s", hdr.Name)
			// make sure we are within a certain path
			if !strings.HasPrefix(abs, path) {
				continue
			}
			log.Println(strings.TrimPrefix(abs, path))
			if hdr.Typeflag == tar.TypeDir {
				files = append(files, File{
					AbsPath: abs,
					RelPath: strings.TrimPrefix(abs, path),
					Dir:     true,
				})
				continue
			}
			var content = &bytes.Buffer{}
			if _, err := io.Copy(content, tr); err != nil {
				return nil, fmt.Errorf("copy config bytes: %w", err)
			}
			files = append(files, File{
				AbsPath: abs,
				RelPath: strings.TrimPrefix(abs, path),
				Content: content.Bytes(),
				Size:    hdr.Size,
			})
		}
	}

	return files, nil
}

// AppendLayerFromFiles returns a new image object containing the appended layer
func AppendLayerFromFiles(img ociv1.Image, files []File) (ociv1.Image, error) {
	var w bytes.Buffer
	tarw := tar.NewWriter(&w)
	for _, f := range files {
		hdr := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     f.AbsPath,
			Size:     f.Size,
		}
		if f.Dir {
			hdr.Typeflag = tar.TypeDir
		}
		log.Printf("name=%s size=%d t=%s", hdr.Name, hdr.Size, string(hdr.Typeflag))

		if err := tarw.WriteHeader(hdr); err != nil {
			return nil, fmt.Errorf("write hdr: %w", err)
		}
		if _, err := tarw.Write(f.Content); err != nil {
			return nil, fmt.Errorf("write file content: %w", err)
		}
	}
	tarw.Close()
	ret, err := mutate.AppendLayers(img, static.NewLayer(w.Bytes(), types.DockerUncompressedLayer))
	if err != nil {
		return nil, fmt.Errorf("append layer: %w", err)
	}
	return ret, nil
}
