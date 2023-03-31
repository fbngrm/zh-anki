package dialog

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fbngrm/nprc/pkg/anki"
	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/hash"
	"github.com/fbngrm/nprc/pkg/ignore"
	"github.com/fbngrm/nprc/pkg/pinyin"
	"github.com/fbngrm/nprc/pkg/sentence"
	"github.com/fbngrm/nprc/pkg/translate"
)

type Dialog struct {
	Deck      string              `yaml:"deck"`
	Tags      []string            `yaml:"tags"`
	Chinese   string              `yaml:"chinese"`
	Pinyin    string              `yaml:"pinyin"`
	English   string              `yaml:"english"`
	Audio     string              `yaml:"audio"`
	Sentences []sentence.Sentence `yaml:"sentences"`
}

type Processor struct {
	Sentences sentence.Processor
	Audio     audio.Downloader
	Exporter  anki.Exporter
}

func (p *Processor) Get(path, deckname string, i ignore.Ignored, pinyinDict pinyin.Dict, t translate.Translations) []*Dialog {
	rawInput := load(path)

	var dialogues []*Dialog
	for _, sentences := range rawInput {
		d := &Dialog{
			Chinese:   strings.Join(sentences, " "),
			Sentences: p.Sentences.Get(sentences, i, pinyinDict, t),
			Deck:      deckname,
		}
		addPinyin(d, pinyinDict)
		translateDialog(d, t)
		p.fetchAudio(d)
		dialogues = append(dialogues, d)
	}
	return dialogues
}

func (p *Processor) fetchAudio(d *Dialog) {
	ctx := context.Background()

	filename, err := p.Audio.Fetch(ctx, d.Chinese, hash.Sha1(d.Chinese))
	if err != nil {
		fmt.Println(err)
	}
	d.Audio = filename
}

func addPinyin(d *Dialog, p pinyin.Dict) {
	pinyin := ""
	for _, sentence := range d.Sentences {
		pinyin += sentence.Pinyin
		pinyin += " "
	}
	d.Pinyin = strings.Trim(pinyin, " ")
}

func translateDialog(d *Dialog, t translate.Translations) {
	translation, ok := t[d.Chinese]
	if !ok {
		var err error
		translation, err = translate.Translate("en-US", d.Chinese)
		if err != nil {
			log.Fatal(err)
		}
	}
	d.English = translation
	t.Update(d.Chinese, d.English)
}

func (p *Processor) Export(dialogues []*Dialog) {
	// for i, dialog := range dialogues {
	// dialogPath := filepath.Join(cwd, "data", deck, lesson, "output", fmt.Sprintf("dialog_%02d.yaml", i+1))
	// writeToFile(dialog, dialogPath)
	// }
}
