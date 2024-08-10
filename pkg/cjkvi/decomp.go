package cjkvi

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fbngrm/zh/pkg/encoding"
)

func NewDecompositionIndex(src string) (map[string][]string, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("could not open ids source file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	components := make(map[string][]string)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		parts := strings.Fields(line)
		if len(line) < 3 {
			continue
		}
		components[parts[1]] = getComponents(parts[2])
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return components, nil
}

func getComponents(ids string) []string {
	// add components from IDS, ignoring the IDS characters
	components := make([]string, 0)
	for _, ideograph := range ids {
		if encoding.IsIdeographicDescriptionCharacter(ideograph) {
			continue
		}
		if ideograph == ' ' {
			continue
		}
		if ideograph == '[' {
			break
		}
		components = append(components, string(ideograph))
	}
	return components
}
