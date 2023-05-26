package frequency

import (
	"bufio"
	"os"
	"strings"

	enc "github.com/fbngrm/zh-anki/pkg/encoding"
)

type WordIndex struct {
	path  string
	Words []string
}

func NewWordIndex(frequencyIndexSrc string) (*WordIndex, error) {
	c := WordIndex{
		path: frequencyIndexSrc,
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (i *WordIndex) init() error {
	file, err := os.Open(i.path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		index = append(index, parts[0])
	}
	i.Words = index

	return scanner.Err()
}

func (wi *WordIndex) GetExamplesForHanzi(hanzi string, count int) []string {
	examples := []string{}
	for _, w := range wi.Words {
		if !strings.Contains(w, hanzi) {
			continue
		}
		examples = append(examples, w)
		if len(examples) == count {
			return examples
		}
	}
	return examples
}

func (wi *WordIndex) GetMostFrequent(limit int) []string {
	known := []string{}
	mostFreq := []string{}
	for i, w := range wi.Words {
		var skip bool
		for _, c := range w {
			// we skip words that contain non-hanzi characters
			if !(enc.DetectRuneType(c) == enc.RuneType_CJKUnifiedIdeograph) {
				skip = true
				break
			}
			if !contains(known, string(c)) {
				mostFreq = append(mostFreq, string(c))
				known = append(known, string(c))
			}
		}
		if !contains(known, w) && !skip {
			mostFreq = append(mostFreq, w)
			known = append(known, w)
		}
		if i == limit {
			break
		}
	}
	return mostFreq
}

func contains(s []string, target string) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}
