package audio

import (
	"io"
	"os"
	"path"

	"golang.org/x/exp/slog"
)

type Cache struct {
	SrcDir string
	DstDir string
}

func (c *Cache) Get(filename string) bool {
	src := path.Join(c.SrcDir, filename)

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return false
	}

	srcFile, err := os.Open(src)
	if err != nil {
		slog.Error("open cached audio file", "err", err)
		return false
	}
	defer srcFile.Close()

	if err := os.MkdirAll(c.DstDir, os.ModePerm); err != nil {
		slog.Error("create tmp audio cache dir", "err", err)
		return false
	}

	dstFile, err := os.Create(path.Join(c.DstDir, filename))
	if err != nil {
		slog.Error("create tmp audio cache file", "err", err)
		return false
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		slog.Error("copy audio cache file", "err", err)
		return false
	}

	return true
}
