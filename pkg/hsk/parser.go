package hsk

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/exp/slog"
)

type Entry struct {
	Ch      string
	Pinyin  string
	Meaning string
	Level   string
}

func clean(s string) string {
	parts := strings.SplitN(s, "｜", 2)
	s = strings.TrimSpace(parts[0])
	parts = strings.Split(s, "（")
	return strings.TrimSpace(parts[0])
}

func NewDict(src string) (map[string]Entry, error) {
	files, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}
	dict := make(map[string]Entry)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(src, file.Name())
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = '\t'

		records, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			if len(record) >= 3 {
				key := clean(record[0])
				value := Entry{
					Ch:      record[0],
					Pinyin:  clean(record[1]),
					Meaning: record[2],
					Level:   strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name())),
				}
				dict[key] = value
			}
		}
	}
	return dict, nil
}

func GetByLevel(dict map[string]Entry, level int) []string {
	all := make(map[string]int)
	byLevel := []string{}
	for k, entry := range dict {
		if entry.Level == strconv.Itoa(level) {
			byLevel = append(byLevel, k)
			if c, ok := all[k]; ok {
				all[k] = c + 1
			} else {
				all[k] = 1
			}
			for _, c := range k {
				cs := string(c)
				if cc, ok := all[cs]; ok {
					all[cs] = cc + 1
				} else {
					all[cs] = 1
				}
			}
		}
	}
	total := 0
	for _, c := range all {
		total += c
	}
	slog.Info(
		"",
		"hsk level", level,
		"distinct words and chars", len(all),
		"expected cards to add", total)
	return byLevel
}
