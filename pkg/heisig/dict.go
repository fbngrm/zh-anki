package heisig

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

type Entry struct {
	SimplifiedChinese  string
	TraditionalChinese string
	Pinyin             string
	Meaning            string
}

func NewDict(sourceFilePath string) (map[string]Entry, error) {
	file, err := os.Open(sourceFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not open heisig dict source file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	decompositions := make(map[string]Entry)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] == '/' {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			slog.Warn(fmt.Sprintf("heisig dict: line too short [len=%d]: %s", len(parts), line))
			continue
		}

		hanzi := strings.Split(parts[0], "[")
		if len(hanzi) < 2 {
			slog.Warn(fmt.Sprintf("heisig dict: hanzi too short [len=%d]: %s", len(hanzi), parts[0]))
			continue
		}
		traditional := strings.TrimSuffix(hanzi[1], "]")

		meaning := parts[3]
		if len(parts) > 4 {
			meaning = strings.Join(parts[3:], " ")
		}

		decompositions[hanzi[0]] = Entry{
			SimplifiedChinese:  hanzi[0],
			TraditionalChinese: traditional,
			Pinyin:             parts[1],
			Meaning:            meaning,
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return decompositions, nil
}
