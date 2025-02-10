package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/segment"
	"golang.org/x/exp/slog"
)

const decomposeDialogMessage = `Add pinyin to the following sentences written in simplified Chinese. Format the result into a JSON object. The original sentence should be stored in a field called chinese, the English translation should be stored in a field called english and the pinyin should be stored in a field called piniyin. Also split each sentence into words and add a JSON array with these words in a field called words. Each word should be a JSON object, the original Chinese word is stored in a field called ch, the English translation is stored in a field called en and the pinyin is stored in a field called pi. For piyin always use the the special characters with accents on top and not the numbers behind the character!`

const sentenceMessage = `Add pinyin to the following sentence written in simplified Chinese. Format the result into a JSON object. The original sentence should be stored in a field called chinese, the English translation should be stored in a field called english and the pinyin should be stored in a field called piniyin. Also split the sentence into words and add JSON array with those words in a field called words. Each word should be a JSON object, the original Chinese word is stored in a field called ch, the English translation is stored in a field called en and the pinyin is stored in a field called pi. For piyin always use the the special characters with accents on top and not the numbers behind the character! Please take extra care, if the input has multiple sentences, do not split them but treat them as asingle sentence. Return a single JSON object only, not a list or array of several ones! Here is the json structure that should be returned:
type Sentence struct {
	Chinese string
	English string
	Pinyin  string
	Words   []Word
}`

const wordExamplesMessage = `Give me three very simple and short Chinese example sentences for the usage of the Chinese word provided by the user. Separate each word in the sentence by a whitespace, this is very important! Use simplified Chinese characters. Also add the pinyin and the english translation. Serialize the response into a JSON object and add the sentences in a JSON array that is referenced by the key "examples". Each example sentence in the array should be a JSON object which has the following fields:
1. "ch": the example sentence in simplified Chinese (each word separated by a whitespace)
2. "pi": the piyin for the example sentence
3. "en": the English translation of the example sentence

Optionally, also add short note to the result if there is anything special to point out on the usage of the word. Maybe there are very similar words which could be confused with the word, or there are common mistakes or misunderstandings that a learner of the Chinese language should be aware of. If the word is frequently used in a certain grammatical context or sentence patterns, please also explain this in the most concise and short manner. Add the note to the response's JSON object in a field called "note". If the note is empty, you do not need to add the field at all. Keep the note as simple and short as possible. Do not add useless information like: "Pay attention to the correct usage of this word in various daily situations." or "Pay attention to the correct order of objects after the word" and the like. We can assume the user always pays attention but wants to know specific details, caveats, casual usages, formal usages, gotchas, common mistakes or hints specific to this word.
`
const patternExamplesMessage = `Give me three very simple Chinese example sentences for the usage of the Chinese grammar pattern provided by the user. Segment the sentences by separating each word in the sentence by a whitespace. Use simplified Chinese characters. Also add the pinyin and the english translation. Serialize the response into a JSON dict and add the sentences in a JSON array that is referenced by the key "examples". Each example sentence in the array should be a JSON dict which has the following fields:
1. "ch": the example sentence in simplified Chinese
2. "pi": the piyin for the example sentence
3. "en": the English translation of the example sentence

Optionally, also add short note to the result if there is anything special to point out on the usage of the sentence pattern. Maybe there are very similar patterns which could be confused with the pattern, or there are common mistakes or misunderstandings that a learner of the Chinese language should be aware of. If the pattern is frequently used in a certain grammatical context, please also explain this in the most concise and short manner. Add the note to the response's JSON dict in a field called "note". If the note is empty, you do not need to add the field at all. Keep the note as simpleand short as possible. Do not add useless information like: "Pay attention to the correct usage of this word in various daily situations." or "Pay attention to the correct order of objects after the word" and the like. We can assume the user always pays attention but wants to know specific details, caveats, casual usages, formal usages, gotchas, common mistakes or hints specific to this word.
`

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Word struct {
	Ch string `json:"ch"`
	En string `json:"en"`
	Pi string `json:"pi"`
}

type ExampleSentences struct {
	Examples []Word `json:"examples"`
	Note     string `json:"note"`
}

type Sentence struct {
	Chinese string `json:"chinese"`
	English string `json:"english"`
	Pinyin  string `json:"pinyin"`
	Words   []Word `json:"words"`
}

type Decomposition struct {
	Sentences []Sentence `json:"sentences"`
}

type Response struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
		Message      struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"message"`
	} `json:"choices"`
	Created float64 `json:"created"`
	ID      string  `json:"id"`
	Model   string  `json:"model"`
	Object  string  `json:"object"`
	Usage   struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Client struct {
	endpoint  string
	apiKey    string
	model     string
	cache     *Cache
	segmenter *segment.Segmenter
}

func NewClient(apiKey string, cache *Cache, segmenter *segment.Segmenter) (*Client, error) {
	if cache == nil {
		return nil, errors.New("cache must not be nil")
	}
	return &Client{
		endpoint:  "https://api.openai.com/v1/chat/completions",
		apiKey:    apiKey,
		model:     "gpt-3.5-turbo",
		cache:     cache,
		segmenter: segmenter,
	}, nil
}

func (c *Client) GetExamplesForPattern(pattern string) (ExampleSentences, error) {
	content := c.fetch(pattern, patternExamplesMessage, 2)

	var result ExampleSentences
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return result, fmt.Errorf("Error parsing JSON for example sentences input %s: %v", content, err)
	}

	// segment chinese text
	examples, err := c.segmentExamples(result.Examples)
	if err != nil {
		slog.Error("Segment word example sentences", "pattern", pattern, "error", err)
	}
	result.Examples = examples

	return result, nil
}

func (c *Client) GetExamplesForWord(word string) (ExampleSentences, error) {
	content := c.fetch(word, wordExamplesMessage, 2)

	var result ExampleSentences
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return result, fmt.Errorf("Error parsing JSON for example sentences input %s: %v", content, err)
	}
	examples, err := c.segmentExamples(result.Examples)
	if err != nil {
		slog.Error("Segment word example sentences", "word", word, "error", err)
	}
	result.Examples = examples
	return result, nil
}

func (c *Client) DecomposeSentence(sentence string) (*Sentence, error) {
	content := c.fetch(sentence, sentenceMessage, 2)
	var result Sentence
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON for sentences input %s: %v", content, err)
	}
	return &result, nil
}

func (c *Client) Decompose(dialog string) (*Decomposition, error) {
	content := c.fetch(dialog, decomposeDialogMessage, 2)

	var sentences []Sentence
	if strings.Contains(content, "\"sentences\": [") {
		var decomp Decomposition
		err := json.Unmarshal([]byte(content), &decomp)
		if err != nil {
			return nil, fmt.Errorf("Error parsing JSON sentences for dialog input %s: %v", content, err)
		}
		sentences = decomp.Sentences
	} else {
		err := json.Unmarshal([]byte(content), &sentences)
		if err != nil {
			log.Printf("Error parsing JSON for dialog input %s: %v", content, err)
		}
	}

	// segment sentences
	words := make([]Word, len(sentences))
	for i, s := range sentences {
		words[i] = Word{Ch: s.Chinese}
	}
	var err error
	words, err = c.segmentExamples(words)
	if err != nil {
		slog.Error("Segment dialog sentences", "dialog", dialog, "error", err)
	}
	// note, we rely on the segmenter func to check boundaries
	for i, w := range words {
		sentences[i].Chinese = w.Ch
	}

	return &Decomposition{
		Sentences: sentences,
	}, nil
}

// FIXME: use filename for cloze cache lookup
// implements a very simple retry. openai api sometimes fails to deliver a result or returns a invalid json
// sub-sequent requests might succeed so we naively just try `retryCount` times.
func (c *Client) fetch(query, message string, retryCount int) string {
	if retryCount == -1 {
		log.Fatalf("excceded retries for query: %s\n", query)
	}
	slog.Info("lookup", "query", query)

	if content, ok := c.cache.Lookup(query); ok {
		slog.Debug("found in cache", "file", query)
		return content
	}
	slog.Debug("not found in cache", "file", query)

	messages := []Message{
		{
			Role:    "system",
			Content: message,
		},
		{
			Role:    "user",
			Content: query,
		},
	}
	payload := Request{
		Model:    c.model,
		Messages: messages,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("error marshalling JSON payload: %v", err)
		fmt.Println("retry...")
		return c.fetch(query, message, retryCount-1)
	}

	// Set up the request object
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("error creating request: %v", err)
		fmt.Println("retry...")
		return c.fetch(query, message, retryCount-1)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error sending request: %v", err)
		fmt.Println("retry...")
		return c.fetch(query, message, retryCount-1)
	}
	defer resp.Body.Close()

	// Read the response body
	// var result map[string]interface{}
	var result Response
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("error decoding JSON response: %v\n", err)
		fmt.Println("retry...")
		return c.fetch(query, message, retryCount-1)
	}

	if len(result.Choices) == 0 {
		fmt.Println("no result, retry...")
		return c.fetch(query, message, retryCount-1)
	}

	content := strings.TrimPrefix(result.Choices[0].Message.Content, "```")
	content = strings.TrimPrefix(result.Choices[0].Message.Content, "json")
	content = strings.TrimSuffix(content, "```")

	c.cache.Add(query, content)

	return content
}

func (c *Client) segmentExamples(in []Word) ([]Word, error) {
	// segment chinese examples
	examples := ""
	for i, e := range in {
		examples += e.Ch
		if i < len(in)-1 {
			examples += "\n"
		}
	}

	if len(examples) > 0 {
		var err error
		examples, err = c.segmenter.SegmentChinese(examples)
		if err != nil {
			return in, err
		}
	}

	segmented := strings.Split(examples, "\n")

	// remove last line if empty
	if len(segmented[len(segmented)-1]) == 0 {
		segmented = segmented[:len(segmented)-1]
	}

	if len(segmented) != len(in) {
		return in, fmt.Errorf("expected %d segmented sentences but got %d", len(in), len(segmented))
	}

	for i, s := range segmented {
		in[i].Ch = s
	}
	return in, nil
}
