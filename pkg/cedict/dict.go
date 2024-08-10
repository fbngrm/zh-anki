package cedict

import (
	"bufio"
	"os"
	"strings"
)

type Entry struct {
	Traditional string
	Simplified  string
	Readings    string
	Definitions []string
}

func NewDict(src string) (map[string][]Entry, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dict := make(map[string][]Entry)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "[")
		if len(parts) == 0 {
			continue
		}

		ideographs := strings.Fields(parts[0])
		if len(ideographs) == 0 {
			continue
		}
		traditional := ideographs[0]
		simplified := ""
		if len(ideographs) > 1 {
			simplified = ideographs[1]
		} else {
			simplified = traditional
		}

		readingsAndDef := strings.Split(parts[1], "]")
		readings := strings.ToLower(readingsAndDef[0])
		definitions := strings.Split(
			strings.Trim(
				strings.TrimSpace(readingsAndDef[1]),
				"/"),
			"/")

		entries, _ := dict[simplified]
		entries = append(entries, Entry{
			Traditional: traditional,
			Simplified:  simplified,
			Readings:    readings,
			Definitions: definitions,
		})
		dict[simplified] = entries
	}
	return dict, scanner.Err()
}
