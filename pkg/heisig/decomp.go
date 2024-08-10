package heisig

import (
	"encoding/json"
	"fmt"
	"os"
)

func NewDecompositionIndex(src string) (map[string][]string, error) {
	lines, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("could not open heisig ids source file: %w", err)
	}
	var data map[string][]string
	err = json.Unmarshal(lines, &data)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal heisig ids: %w", err)
	}
	return data, nil
}
