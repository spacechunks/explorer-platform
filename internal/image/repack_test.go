package image_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/spacechunks/platform/internal/image"
	"github.com/spacechunks/platform/internal/image/testdata"
	"github.com/stretchr/testify/assert"
)

func TestRepack(t *testing.T) {
	testImgOpener := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(testdata.RepackImage)), nil
	}
	src, err := tarball.Image(testImgOpener, nil)
	if err != nil {
		t.Fatalf("read img: %v", err)
	}
	expected := []image.File{
		{
			AbsPath:      "a/",
			StrippedPath: "/",
			Dir:          true,
		},
		{
			AbsPath:      "a/file1",
			StrippedPath: "/file1",
			Content:      []byte("file1 repacked\n"),
			Size:         15,
		},
		{
			AbsPath:      "a/b/",
			StrippedPath: "/b/",
			Dir:          true,
		},
		{
			AbsPath:      "a/b/file2",
			StrippedPath: "/b/file2",
			Content:      []byte("file2 repacked\n"),
			Size:         15,
		},
		{
			AbsPath:      "a/b/file4",
			StrippedPath: "/b/file4",
			Content:      []byte("file4 repacked\n"),
			Size:         15,
		},
		{
			AbsPath:      "a/b/c/",
			StrippedPath: "/b/c/",
			Dir:          true,
		},
		{
			AbsPath:      "a/b/c/file3",
			StrippedPath: "/b/c/file3",
			Content:      []byte("abc\n"),
			Size:         4,
		},
		{
			AbsPath:      "a/d/",
			StrippedPath: "/d/",
			Dir:          true,
		},
		{
			AbsPath:      "a/d/file",
			StrippedPath: "/d/file",
			Content:      []byte("repacked file\n"),
			Size:         14,
		},
	}

	img, err := image.Repack(src, "a", "overlay")
	if err != nil {
		t.Fatalf("repack: %v", err)
	}

	files, err := image.UnpackDir(img, "a")
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}
	assert.ElementsMatch(t, expected, files)
}
