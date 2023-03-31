package pinyin

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Dict map[string]string

func (p Dict) Update(ch, pi string) {
	p[ch] = pi
}

func Load(path string) Dict {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not open Pinyin file: %v", err)
		os.Exit(1)
	}
	var p Dict
	if err := yaml.Unmarshal(b, &p); err != nil {
		fmt.Printf("could not unmarshal Pinyin file: %v", err)
		os.Exit(1)
	}
	if p == nil {
		p = make(map[string]string)
	}
	return p
}

func (p Dict) Write(path string) {
	data, err := yaml.Marshal(p)
	if err != nil {
		fmt.Printf("could not marshal Pinyin file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write Pinyin file: %v", err)
		os.Exit(1)
	}
}
