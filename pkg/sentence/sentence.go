package sentence

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/hash"
	"github.com/fbngrm/nprc/pkg/ignore"
	pinyin_dict "github.com/fbngrm/nprc/pkg/pinyin"
	"github.com/fbngrm/nprc/pkg/translate"
	"github.com/fbngrm/nprc/pkg/word"
	"github.com/fbngrm/zh/lib/cedict"
)

type Sentence struct {
	Chinese      string      `yaml:"chinese"`
	Pinyin       string      `yaml:"pinyin"`
	English      string      `yaml:"english"`
	Audio        string      `yaml:"audio"`
	NewWords     []word.Word `yaml:"newWords"`
	AllWords     []word.Word `yaml:"allWords"`
	IsSingleRune bool        `yaml:"isSingleRune"`
}

type Processor struct {
	Words  word.Processor
	Cedict map[string][]cedict.Entry
	Audio  audio.Downloader
}

func (p *Processor) Get(sentences []string, i ignore.Ignored, pinyin pinyin_dict.Dict, t translate.Translations) []Sentence {
	var results []Sentence
	for _, sentence := range sentences {
		if sentence == "" {
			continue
		}
		allWords, newWords := p.Words.Get(sentence, i, t)
		results = append(results, Sentence{
			Chinese:      sentence,
			Audio:        hash.Sha1(sentence),
			AllWords:     allWords,
			NewWords:     newWords,
			IsSingleRune: utf8.RuneCountInString(sentence) == 1,
		})
	}
	return p.getAudio(translateSentences(p.addPinyin(results, pinyin), t))
}

// a word can have several pinyins. but in a sentence a word can only have one, so we need to
// find it, in the worst case we have to ask the user.
func (p *Processor) addPinyin(sentences []Sentence, pinyinDict pinyin_dict.Dict) []Sentence {
	for i, sentence := range sentences {
		pinyin := ""
		if p, ok := pinyinDict[sentence.Chinese]; ok {
			sentences[i].Pinyin = p
			continue
		}
		for _, w := range sentence.AllWords {
			if p, ok := pinyinDict[w.Chinese]; ok {
				pinyin += p
				pinyin += " "
				continue
			}
			entries, _ := p.Cedict[w.Chinese]
			allReadings := make(map[string]struct{})
			for _, entry := range entries {
				for _, reading := range entry.Readings {
					allReadings[reading] = struct{}{}
				}
			}
			readings := make([]string, 0)
			for reading := range allReadings {
				readings = append(readings, reading)
			}
			if len(readings) == utf8.RuneCountInString(w.Chinese) {
				pinyin += strings.Join(readings, "")
				pinyin += " "
				continue
			}
			if len(readings) == 0 {
				fmt.Println("===============================================")
				fmt.Printf("sentence: %s\n", sentence.Chinese)
				fmt.Printf("no readings found for word \"%s\", please enter pinyin\n", w.Chinese)
				pinyin += getPinyinFromUser(sentence.Chinese, nil)
				pinyin += " "
				continue
			}
			if len(readings) > 1 {
				fmt.Println("===============================================")
				fmt.Printf("sentence: %s\n", sentence.Chinese)
				fmt.Printf("more than 1 readings found for word \"%s\" please choose\n", w.Chinese)
				pinyin += getPinyinFromUser(sentence.Chinese, readings)
				pinyin += " "
				continue
			}
			if len(readings) == 1 {
				pinyin += readings[0]
				pinyin += " "
				continue
			}
		}
		r, _ := utf8.DecodeLastRuneInString(sentence.Chinese)
		p := strings.Trim(pinyin, " ")
		p += string(r)
		sentences[i].Pinyin = p
		pinyinDict.Update(sentence.Chinese, p)
	}

	return sentences
}

func (p *Processor) getAudio(sentences []Sentence) []Sentence {
	for x, sentence := range sentences {
		filename, err := p.Audio.Fetch(context.Background(), sentence.Chinese, hash.Sha1(sentence.Chinese))
		if err != nil {
			fmt.Println(err)
		}
		sentences[x].Audio = filename
	}
	return sentences
}

func getPinyinFromUser(sentence string, options []string) string {
	scanner := bufio.NewScanner(os.Stdin)
	if len(options) > 1 {
		for i, o := range options {
			fmt.Printf("option %d: %s\n", i+1, o)
		}
		scanner.Scan()
		key := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Printf("could not read input: %v\n", err)
			os.Exit(1)
		}
		i, err := strconv.Atoi(key)
		if err != nil {
			fmt.Println(err)
			fmt.Println("invalid index")
			return getPinyinFromUser(sentence, options)
		}
		if i-1 < 0 || i-1 >= len(options) {
			fmt.Println("invalid index")
			return getPinyinFromUser(sentence, options)
		}
		return options[i-1]
	} else {
		scanner.Scan()
		text := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Printf("could not read input: %v\n", err)
			os.Exit(1)
		}
		return text
	}
}

func translateSentences(sentences []Sentence, t translate.Translations) []Sentence {
	for i, sentence := range sentences {
		translation, ok := t[sentence.Chinese]
		if !ok {
			var err error
			translation, err = translate.Translate("en-US", sentence.Chinese)
			if err != nil {
				log.Fatalf("could not translate sentence \"%s\": %v", sentence.Chinese, err)
			}
		}
		sentence.English = translation
		t.Update(sentence.Chinese, sentence.English)
		sentences[i] = sentence
	}
	return sentences
}
