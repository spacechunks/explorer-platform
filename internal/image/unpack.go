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
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"time"
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

func UnpackFile(img ociv1.Image, path string) (File, error) {
	layers, err := img.Layers()
	if err != nil {
		return File{}, fmt.Errorf("image layers: %w", err)
	}

	// we will find our file most likely in
	// the latest layers that have been added.
	// new layers are appended to the end of the
	// slice, so we have to reverse it.
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

// UnpackDir applies each layers change set for a given directory.
// Returns all files living under the specified path.
func UnpackDir(img ociv1.Image, path string) (map[string]File, error) {
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("image layers: %w", err)
	}
	var (
		g            = &errgroup.Group{}
		uncompressed = make([]io.ReadCloser, len(layers))
	)

	t1 := time.Now()

	for i, l := range layers {
		i, l := i, l
		g.Go(func() error {
			data, err := l.Uncompressed()
			if err != nil {
				return err
			}
			uncompressed[i] = data
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("uncompress layers: %v", err)
	}

	t2 := time.Now()

	log.Printf("uncompress: %v\n", t2.Sub(t1))

	t3 := time.Now()

	// fmap stores the file hierarchy
	// mapping path to file object
	fmap := make(map[string]File)
	// iterate over each uncompressed layer
	// oldest to latest while applying the
	// layers change set
	for _, data := range uncompressed {
		tr := tar.NewReader(data)
		defer data.Close()
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			// tar deals with relative paths, so make em absolute
			abs := fmt.Sprintf("/%s", hdr.Name)
			// make sure we are within path
			if !strings.HasPrefix(abs, path) {
				continue
			}

			// TODO: cut away root dir, because we want

			// files which should be removed are prefixed with .wh
			// see: https://github.com/opencontainers/image-spec/blob/56fb7838abe52ee259e37ece4b314c08bd45997f/layer.md#L246
			//
			// filename := .wh.myscript.sh
			filename := filepath.Base(abs)
			if strings.HasPrefix(filename, ".wh") {
				var (
					// filename := myscript.sh
					f   = filename[4:]
					dir = filepath.Dir(abs)
					// remove all files under p
					rmAll = func(m map[string]File, p string) {
						for k := range fmap {
							if !strings.HasPrefix(k, p) {
								continue
							}
							delete(fmap, k)
						}
					}
				)
				// this indicates that the all files
				// contained within the directory should be deleted.
				// example: /dir1/dir2/.wh..wh..opq
				// see https://github.com/opencontainers/image-spec/blob/56fb7838abe52ee259e37ece4b314c08bd45997f/layer.md#L284
				if filename == ".wh..opq" {
					rmAll(fmap, dir)
					continue
				}
				// path := <path-to-dir>/myscript.sh
				path := fmt.Sprintf("%s/%s", filepath.Dir(abs), f)
				// .wh prefixed directories should also be removed completely
				if hdr.Typeflag == tar.TypeDir {
					rmAll(fmap, path)
					continue
				}
				delete(fmap, path)
				continue
			}
			if hdr.Typeflag == tar.TypeDir {
				fmap[abs] = File{
					AbsPath: abs,
					RelPath: hdr.Name,
					Dir:     true,
				}
				continue
			}
			var content = &bytes.Buffer{}
			if _, err := io.Copy(content, tr); err != nil {
				return nil, fmt.Errorf("copy config bytes: %w", err)
			}
			fmap[abs] = File{
				AbsPath: abs,
				RelPath: hdr.Name,
				Content: content.Bytes(),
				Size:    hdr.Size,
			}
		}
	}

	t4 := time.Now()
	log.Printf("unpack: %v\n", t4.Sub(t3))
	return fmap, nil
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

func applyChangeSet() {

}
