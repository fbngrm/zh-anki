package dialog

import (
	"fmt"
	"os"
	"strings"
)

func loadSimpleGrammar(path string) (SimpleGrammar, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return SimpleGrammar{}, fmt.Errorf("error reading simple grammar file: %v", err)
	}
	s := string(b)
	parts := strings.Split(s, "---")
	if len(parts) != 2 {
		return SimpleGrammar{}, fmt.Errorf("error reading simple grammar file, expect two parts but got: %d", len(parts))
	}
	return SimpleGrammar{
		Head: parts[0],
		HTML: parts[1],
	}, nil
}
