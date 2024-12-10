/*
Explorer Platform, a platform for hosting and discovering Minecraft servers.
Copyright (C) 2024 Yannic Rieger <oss@76k.io>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package image

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"

	ociv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"golang.org/x/sync/errgroup"
)

// TODO: tests

type File struct {
	AbsPath string
	// StrippedPath is the absolute path of
	// the file or directory without its
	// root directory. for path /a/b/file.txt with
	// a being the root directory StrippedPath
	// would be /b/file.txt
	StrippedPath string
	Content      []byte
	Size         int64
	Dir          bool
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
			if hdr.Name != path {
				continue
			}
			var content = &bytes.Buffer{}
			if _, err := io.Copy(content, tr); err != nil {
				return File{}, fmt.Errorf("copy file bytes: %w", err)
			}
			return File{
				AbsPath:      hdr.Name,
				StrippedPath: strings.TrimPrefix(hdr.Name, path),
				Content:      content.Bytes(),
				Size:         hdr.Size,
			}, nil
		}
	}

	return File{}, errors.New("file not found")
}

// UnpackDir applies each layers change set for a given directory.
// Returns all files living under the specified path.
func UnpackDir(img ociv1.Image, path string) ([]File, error) {
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("image layers: %w", err)
	}
	uncompressed, err := cmap(layers, func(l ociv1.Layer) (io.ReadCloser, error) {
		data, err := l.Uncompressed()
		if err != nil {
			return nil, err
		}
		return data, nil
	})
	if err != nil {
		return nil, fmt.Errorf("uncompress layers: %v", err)
	}

	// iterate over each uncompressed layer
	// oldest to latest reading the change sets
	changeSets, err := cmap(uncompressed, func(data io.ReadCloser) (map[string]File, error) {
		// fmap stores the file hierarchy
		// mapping path to file object
		fmap := make(map[string]File)
		tr := tar.NewReader(data)
		defer data.Close()
		//t1 := time.Now()
		//log.Printf("starting layer: %s\n", t1.String())
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			// make sure we are within path
			if !strings.HasPrefix(hdr.Name, path) {
				continue
			}
			if hdr.Typeflag == tar.TypeDir {
				fmap[hdr.Name] = File{
					AbsPath:      hdr.Name,
					StrippedPath: strings.TrimPrefix(hdr.Name, path),
					Dir:          true,
				}
				continue
			}

			var content = &bytes.Buffer{}
			if _, err := io.Copy(content, tr); err != nil {
				return nil, fmt.Errorf("copy config bytes: %w", err)
			}
			fmap[hdr.Name] = File{
				AbsPath:      hdr.Name,
				StrippedPath: strings.TrimPrefix(hdr.Name, path),
				Content:      content.Bytes(),
				Size:         hdr.Size,
			}
		}
		//t2 := time.Now()
		//log.Printf("layer: %v\n", t2.Sub(t1))
		return fmap, nil
	})
	if err != nil {
		return nil, fmt.Errorf("change sets: %v", err)
	}

	final := make(map[string]File)
	for _, cs := range changeSets {
		for abs, file := range cs {
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
						for k := range m {
							if !strings.HasPrefix(k, p) {
								continue
							}
							delete(m, k)
						}
					}
				)
				// this indicates that the all files
				// contained within the directory should be deleted.
				// example: /dir1/dir2/.wh..wh..opq
				// see https://github.com/opencontainers/image-spec/blob/56fb7838abe52ee259e37ece4b314c08bd45997f/layer.md#L284
				if filename == ".wh..opq" {
					rmAll(final, dir)
					continue
				}
				// .wh prefixed directories should also be removed completely
				if file.Dir {
					rmAll(final, dir)
					continue
				}
				// path := <path-to-dir>/myscript.sh
				delete(final, fmt.Sprintf("%s/%s", filepath.Dir(abs), f))
				continue
			}
			final[abs] = file
		}
	}
	return values(final), nil
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
	if err := tarw.Close(); err != nil {
		return nil, fmt.Errorf("tar close: %w", err)
	}
	ret, err := mutate.AppendLayers(img, static.NewLayer(w.Bytes(), types.DockerUncompressedLayer))
	if err != nil {
		return nil, fmt.Errorf("append layer: %w", err)
	}
	return ret, nil
}

// cmap maps each element concurrently using an errgroup.Group.
// returns a slice containing the mapped elements.
func cmap[T any, K any](in []K, fn func(K) (T, error)) ([]T, error) {
	g := &errgroup.Group{}
	s := make([]T, len(in))
	for i, val := range in {
		i, val := i, val
		g.Go(func() error {
			ret, err := fn(val)
			if err != nil {
				return err
			}
			s[i] = ret
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("cmap: %v", err)
	}
	return s, nil
}

func values(m map[string]File) []File {
	s := make([]File, 0, len(m))
	for _, f := range m {
		s = append(s, f)
	}
	return s
}
