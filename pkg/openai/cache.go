package openai

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/hash"
)

type Cache struct {
	dir      string
	index    map[string]struct{}
	exporter *anki.Exporter
}

func NewCache(dir string, exporter *anki.Exporter) *Cache {
	return &Cache{
		dir:      dir,
		index:    read(dir),
		exporter: exporter,
	}
}

func (c *Cache) Lookup(key string) (string, bool) {
	filename := hash.Sha1(key) + ".yaml"
	if _, ok := c.index[filename]; ok {
		f, err := os.Open(filepath.Join(c.dir, filename))
		if err != nil {
			log.Printf("Failed to open file %s: %v", filename, err)
			return "", false
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filename, err)
			return "", false
		}
		return string(data), true
	}
	return "", false
}

func (c *Cache) Add(key, val string) {
	key = hash.Sha1(key) + ".yaml"
	path := filepath.Join(c.dir, key)
	if err := os.WriteFile(path, []byte(val), 0644); err != nil {
		fmt.Printf("could not write cache file: %v", err)
		os.Exit(1)
	}
	c.index[key] = struct{}{}
}

func read(dir string) map[string]struct{} {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return map[string]struct{}{}
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	filenames := make(map[string]struct{})
	for _, file := range files {
		filenames[file.Name()] = struct{}{}
	}
	return filenames
}
