package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

var in string
var templatePath string
var deckName string

func main() {
	flag.StringVar(&in, "i", "", "raw input")
	flag.StringVar(&deckName, "d", "", "deck name")
	flag.Parse()

	ignorePath := filepath.Join(in, "..", "..", "ignore")
	ignored := loadIgnored(ignorePath)
	sentences := loadSentences(in)
	dialog := Dialog{
		Chinese:   strings.Join(sentences, " "),
		Sentences: getSentences(ignored, sentences),
		Deck:      deckName,
		Tags:      []string{"hsk1", "new-practical-chinese-reader-01"},
	}

	translationsPath := filepath.Join(in, "..", "..", "translations")
	translations := loadTranslations(translationsPath)
	translateNew(translations, &dialog)
	translateAll(translations, &dialog)

	audioDirPath := filepath.Join(in, "..", "audio")
	getAudio(audioDirPath, &dialog)

	dialogPath := filepath.Join(in, "..", "dialog.yaml")
	writeDialog(dialog, dialogPath)
	writeIgnored(ignored, ignorePath)
	writeTranslations(translations, translationsPath)

	templatePath = filepath.Join(in, "..", "..", "tmpl", deckName+".tmpl")
	outPath := filepath.Join(in, "..", "cards.md")
	exportSentences(&dialog, deckName, templatePath, outPath)
}

// parse

type Ignored map[string]struct{}

func (i Ignored) update(s string) {
	i[s] = struct{}{}
}

type Char struct {
	Chinese string `yaml:"chinese"`
	English string `yaml:"english"`
}

type Word struct {
	Chinese  string `yaml:"chinese"`
	English  string `yaml:"english"`
	Audio    string `yaml:"audio"`
	NewChars []Char `yaml:"newChars"`
	AllChars []Char `yaml:"allChars"`
}

type Sentence struct {
	Chinese  string `yaml:"chinese"`
	English  string `yaml:"english"`
	Audio    string `yaml:"audio"`
	NewWords []Word `yaml:"newWords"`
	AllWords []Word `yaml:"allWords"`
}

type Dialog struct {
	Deck      string     `yaml:"deck"`
	Tags      []string   `yaml:"tags"`
	Chinese   string     `yaml:"chinese"`
	English   string     `yaml:"english"`
	Audio     string     `yaml:"audio"`
	Sentences []Sentence `yaml:"sentences"`
}

func getSentences(ignore Ignored, sentences []string) []Sentence {
	var results []Sentence
	for _, sentence := range sentences {
		if sentence == "" {
			continue
		}
		allWords, newWords := getWords(ignore, sentence)
		results = append(results, Sentence{
			Chinese:  sentence,
			Audio:    hash(sentence),
			AllWords: allWords,
			NewWords: newWords,
		})
	}
	return results
}

func getWords(ignore Ignored, sentence string) ([]Word, []Word) {
	sentence = strings.ReplaceAll(sentence, "。", " ")
	sentence = strings.ReplaceAll(sentence, ".", " ")
	sentence = strings.ReplaceAll(sentence, "，", " ")
	sentence = strings.ReplaceAll(sentence, ",", " ")
	sentence = strings.ReplaceAll(sentence, "?", " ")
	sentence = strings.ReplaceAll(sentence, "？", " ")
	sentence = strings.ReplaceAll(sentence, "！", " ")
	sentence = strings.ReplaceAll(sentence, "!", " ")

	var parts []string
	if strings.Contains(sentence, " ") {
		parts = strings.Split(sentence, " ")
	} else if strings.Contains(sentence, " ") {
		parts = strings.Split(sentence, " ")
	} else {
		parts = []string{sentence}
	}

	var allWords []Word
	for _, word := range parts {
		if word == "" {
			continue
		}
		allChars := getAllChars(word)
		allWords = append(allWords, Word{
			Chinese:  word,
			Audio:    hash(word),
			AllChars: allChars,
		})
	}

	var newWords []Word
	for _, word := range allWords {
		if _, ok := ignore[word.Chinese]; ok {
			continue
		}
		ignore.update(word.Chinese)

		// set new chars after word has been added to ignore list,
		// we want to add words first, then chars
		word.NewChars = getNewChars(ignore, word.AllChars)
		newWords = append(newWords, word)
	}

	return allWords, newWords
}

func getAllChars(word string) []Char {
	allChars := make([]Char, 0)
	for _, char := range word {
		allChars = append(allChars, Char{
			Chinese: string(char),
		})
	}
	return allChars
}

func getNewChars(ignore Ignored, allChars []Char) []Char {
	newChars := make([]Char, 0)
	for _, char := range allChars {
		if _, ok := ignore[char.Chinese]; ok {
			continue
		}
		newChars = append(newChars, char)
		ignore.update(char.Chinese)
	}
	return newChars
}

func loadSentences(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open dialogues file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == " " {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func loadIgnored(path string) Ignored {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not open ignore file: %v", err)
		os.Exit(1)
	}
	var i Ignored
	if err := yaml.Unmarshal(b, &i); err != nil {
		fmt.Printf("could not unmarshal ignore file: %v", err)
		os.Exit(1)
	}
	return i
}

func writeIgnored(i Ignored, path string) {
	data, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("could not marshal ignore file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write ignore file: %v", err)
		os.Exit(1)
	}
}

func writeDialog(d Dialog, path string) {
	data, err := yaml.Marshal(d)
	if err != nil {
		fmt.Printf("could not marshal dialog file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write dialog file: %v", err)
		os.Exit(1)
	}
}

// translate

type Translations map[string]string

func (t Translations) update(ch, en string) {
	t[ch] = en
}

func translateNew(t Translations, d *Dialog) {
	translation, ok := t[d.Chinese]
	if !ok {
		var err error
		translation, err = translateText("en-US", d.Chinese)
		if err != nil {
			fmt.Println(err)
		}
	}
	d.English = translation
	t.update(d.Chinese, d.English)

	for i, sentence := range d.Sentences {
		translation, ok := t[sentence.Chinese]
		if !ok {
			var err error
			translation, err = translateText("en-US", sentence.Chinese)
			if err != nil {
				log.Fatalf("could not translate sentence \"%s\": %v", sentence.Chinese, err)
			}
		}
		sentence.English = translation
		t.update(sentence.Chinese, sentence.English)

		for y, word := range sentence.NewWords {
			translation, ok := t[word.Chinese]
			if !ok {
				var err error
				translation, err = translateText("en-US", word.Chinese)
				if err != nil {
					log.Fatalf("could not translate word \"%s\": %v", word.Chinese, err)
				}
			}
			word.English = translation
			t.update(word.Chinese, word.English)

			for z, char := range word.NewChars {
				translation, ok := t[char.Chinese]
				if !ok {
					var err error
					translation, err = translateText("en-US", char.Chinese)
					if err != nil {
						log.Fatalf("could not translate char \"%s\": %v", char.Chinese, err)
					}
				}
				char.English = translation
				t.update(char.Chinese, char.English)
				word.NewChars[z] = char
			}
			sentence.NewWords[y] = word
		}
		d.Sentences[i] = sentence
	}
}

func translateAll(t Translations, d *Dialog) {
	for i, sentence := range d.Sentences {
		for y, word := range sentence.AllWords {
			translation, ok := t[word.Chinese]
			if !ok {
				var err error
				translation, err = translateText("en-US", word.Chinese)
				if err != nil {
					log.Fatalf("could not translate word \"%s\": %v", word.Chinese, err)
				}
			}
			word.English = translation
			t.update(word.Chinese, word.English)

			for z, char := range word.AllChars {
				translation, ok := t[char.Chinese]
				if !ok {
					var err error
					translation, err = translateText("en-US", char.Chinese)
					if err != nil {
						log.Fatalf("could not translate char \"%s\": %v", char.Chinese, err)
					}
				}
				char.English = translation
				t.update(char.Chinese, char.English)
				word.AllChars[z] = char
			}
			sentence.AllWords[y] = word
		}
		d.Sentences[i] = sentence
	}
}

func translateText(targetLanguage, text string) (string, error) {
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("Translate: %v", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("Translate returned empty response to text: %s", text)
	}
	fmt.Printf("translate: %s...\n", text)
	fmt.Println(resp[0].Text)
	return resp[0].Text, nil
}

func loadTranslations(path string) Translations {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not open translations file: %v", err)
		os.Exit(1)
	}
	var t Translations
	if err := yaml.Unmarshal(b, &t); err != nil {
		fmt.Printf("could not unmarshal translations file: %v", err)
		os.Exit(1)
	}
	return t
}

func writeTranslations(t Translations, path string) {
	data, err := yaml.Marshal(t)
	if err != nil {
		fmt.Printf("could not marshal translations file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write translations file: %v", err)
		os.Exit(1)
	}
}

// audio

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(h.Sum(nil))
}

func getAudio(audioDirPath string, d *Dialog) {
	ctx := context.Background()

	filename, err := fetchAudio(ctx, d.Chinese, audioDirPath, hash(d.Chinese))
	if err != nil {
		fmt.Println(err)
	}
	d.Audio = filename

	for i, sentence := range d.Sentences {
		filename, err := fetchAudio(ctx, sentence.Chinese, audioDirPath, hash(sentence.Chinese))
		if err != nil {
			fmt.Println(err)
		}
		sentence.Audio = filename

		for y, word := range sentence.NewWords {
			filename, err := fetchAudio(ctx, word.Chinese, audioDirPath, hash(word.Chinese))
			if err != nil {
				fmt.Println(err)
			}
			word.Audio = filename
			sentence.NewWords[y] = word
		}
		d.Sentences[i] = sentence
	}
}

func fetchAudio(ctx context.Context, query, dir, filename string) (string, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	filename = filename + ".mp3"
	path := filepath.Join(dir, filename)

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("audio file exists: %s\n", path)
		return filename, nil
	}

	time.Sleep(1 * time.Second)
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: query},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "cmn-CN",
			Name:         "cmn-CN-Wavenet-C",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return "", err
	}

	// The resp's AudioContent is binary.
	err = ioutil.WriteFile(path, resp.AudioContent, 0644)
	if err != nil {
		return "", err
	}

	fmt.Printf("%v\n", query)
	fmt.Printf("audio content written to file: %v\n", path)
	return filename, nil
}

// export

func exportSentences(d *Dialog, deckName, templatePath, outPath string) {
	formatted, err := fillTemplate(d, deckName, templatePath)
	if err != nil {
		fmt.Printf("could not format hanzi: %v\n", err)
		os.Exit(1)
	}
	writeFile(formatted, outPath)
}

func writeFile(data, outPath string) {
	if err := os.WriteFile(outPath, []byte(data), 0644); err != nil {
		fmt.Printf("could not write anki cards: %v", err)
		os.Exit(1)
	}
}

func fillTemplate(d *Dialog, deckName, tmplPath string) (string, error) {
	tplFuncMap := make(template.FuncMap)
	tplFuncMap["audio"] = func(query string) string {
		return "[sound:" + hash(query) + ".mp3]"
	}
	tplFuncMap["removeSpaces"] = func(s string) string {
		return strings.ReplaceAll(s, " ", "")
	}
	tmpl, err := template.New(deckName + ".tmpl").Funcs(tplFuncMap).ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, d)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
