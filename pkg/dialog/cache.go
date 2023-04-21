package dialog

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/nprc/pkg/anki"
	"github.com/fbngrm/nprc/pkg/hash"
	"gopkg.in/yaml.v2"
)

type Cache struct {
	dir       string
	dialogues map[string]struct{}
	sentences map[string]struct{}
	words     map[string]struct{}
	exporter  anki.Exporter
}

func NewCache(dir string) *Cache {
	return &Cache{
		dir:       dir,
		dialogues: read(filepath.Join(dir, "dialogues")),
		sentences: read(filepath.Join(dir, "sentences")),
		words:     read(filepath.Join(dir, "words")),
	}
}

func (c *Cache) lookupDialog(key string) (*Dialog, bool) {
	filename := hash.Sha1(key)
	if _, ok := c.dialogues[filename]; ok {
		var d *Dialog
		f, err := os.Open(filepath.Join(c.dir, filename))
		if err != nil {
			log.Printf("Failed to open file %s: %v", filename, err)
			return nil, false
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filename, err)
			return nil, false
		}
		if err := yaml.Unmarshal(data, &d); err != nil {
			log.Printf("Failed to unmarshal YAML data for dialog %s: %v", filename, err)
			return nil, false
		}
		return d, true
	}
	return nil, false
}

func (c *Cache) lookupSentence(key string) (*Sentence, bool) {
	filename := hash.Sha1(key)
	if _, ok := c.sentences[filename]; ok {
		var s *Sentence
		f, err := os.Open(filepath.Join(c.dir, filename))
		if err != nil {
			log.Printf("Failed to open file %s: %v", filename, err)
			return nil, false
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filename, err)
			return nil, false
		}
		if err := yaml.Unmarshal(data, &s); err != nil {
			log.Printf("Failed to unmarshal YAML data for sentence %s: %v", filename, err)
			return nil, false
		}
		return s, true
	}
	return nil, false
}

func (c *Cache) lookupWord(key string) (bool, *Word) {
	filename := hash.Sha1(key)
	if _, ok := c.words[filename]; ok {
		var w *Word
		f, err := os.Open(filepath.Join(c.dir, filename))
		if err != nil {
			log.Printf("Failed to open file %s: %v", filename, err)
			return false, nil
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filename, err)
			return false, nil
		}
		if err := yaml.Unmarshal(data, &w); err != nil {
			log.Printf("Failed to unmarshal YAML data for word %s: %v", filename, err)
			return false, nil
		}
		return true, w
	}
	return false, nil
}

func (c *Cache) AddSentence(s *Sentence) {
	key := hash.Sha1(s.Chinese) + ".yaml"
	path := filepath.Join(c.dir, key)
	c.exporter.WriteYAMLFile(s, path)
	c.sentences[key] = struct{}{}
}

func (c *Cache) AddDialog(d *Dialog) {
	key := hash.Sha1(d.Chinese) + ".yaml"
	path := filepath.Join(c.dir, key)
	c.exporter.WriteYAMLFile(d, path)
	c.dialogues[key] = struct{}{}
}

func (c *Cache) AddWord(w *Word) {
	key := hash.Sha1(w.Chinese) + ".yaml"
	path := filepath.Join(c.dir, key)
	c.exporter.WriteYAMLFile(w, path)
	c.words[key] = struct{}{}
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
