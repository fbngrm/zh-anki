package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/hash"
	"golang.org/x/exp/slog"
)

const dialogSystemMessage = `Add pinyin to the following sentences written in simplified Chinese. Format the result into a JSON object. The original sentence should be stored in a field called chinese, the English translation should be stored in a field called english and the pinyin should be stored in a field called piniyin. Also split each sentence into words and add a JSON array with these words in a field called words. Each word should be a JSON object, the original Chinese word is stored in a field called ch, the English translation is stored in a field called en and the pinyin is stored in a field called pi. For piyin always use the the special characters with accents on top and not the numbers behind the character!`

const sentenceSystemMessage = `Add pinyin to the following sentence written in simplified Chinese. Format the result into a JSON object. The original sentence should be stored in a field called chinese, the English translation should be stored in a field called english and the pinyin should be stored in a field called piniyin. Also split the sentence into words and add JSON array with those words in a field called words. Each word should be a JSON object, the original Chinese word is stored in a field called ch, the English translation is stored in a field called en and the pinyin is stored in a field called pi. For piyin always use the the special characters with accents on top and not the numbers behind the character! Please take extra care, if the input has multiple sentences, do not split them but treat them as asingle sentence. Return a single JSON object only, not a list or array of several ones! Here is the json structure that should be returned:
type Sentence struct {
	Chinese string
	English string
	Pinyin  string
	Words   []Word
}`

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
	endpoint string
	apiKey   string
	model    string
	cache    *Cache
}

func NewClient(apiKey string, cache *Cache) *Client {
	return &Client{
		endpoint: "https://api.openai.com/v1/chat/completions",
		apiKey:   apiKey,
		model:    "gpt-3.5-turbo",
		cache:    cache,
	}
}

func (c *Client) DecomposeSentence(sentence string) (*Sentence, error) {
	content := c.fetch(sentence, 2)
	var result Sentence
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON for sentences input %s: %v", content, err)
	}
	return &result, nil
}

func (c *Client) Decompose(dialog string) (*Decomposition, error) {
	content := c.fetch(dialog, 2)

	if strings.Contains(content, "\"sentences\": [") {
		var decomp Decomposition
		err := json.Unmarshal([]byte(content), &decomp)
		if err != nil {
			return nil, fmt.Errorf("Error parsing JSON sentences for dialog input %s: %v", content, err)
		}
		return &decomp, nil
	}

	var sentences []Sentence
	err := json.Unmarshal([]byte(content), &sentences)
	if err != nil {
		log.Printf("Error parsing JSON for dialog input %s: %v", content, err)
	}
	return &Decomposition{
		Sentences: sentences,
	}, nil
}

// implements a very simple retry. openai api sometimes fails to deliver a result or returns a invalid json
// sub-sequent requests might succeed so we naively just try `retryCount` times.
func (c *Client) fetch(query string, retryCount int) string {
	if retryCount == -1 {
		log.Fatalf("excceded retries for query: %s\n", query)
	}
	slog.Info("lookup", "query", query)
	if content, ok := c.cache.Lookup(query); ok {
		slog.Debug("found in cache", "file", hash.Sha1(query))
		return content
	}
	slog.Debug("not found in cache", "file", hash.Sha1(query))

	messages := []Message{
		{
			Role:    "system",
			Content: dialogSystemMessage,
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
		return c.fetch(query, retryCount-1)
	}

	// Set up the request object
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("error creating request: %v", err)
		fmt.Println("retry...")
		return c.fetch(query, retryCount-1)
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
		return c.fetch(query, retryCount-1)
	}
	defer resp.Body.Close()

	// Read the response body
	// var result map[string]interface{}
	var result Response
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("error decoding JSON response: %v\n", err)
		fmt.Println("retry...")
		return c.fetch(query, retryCount-1)
	}

	if len(result.Choices) == 0 {
		fmt.Println("no result, retry...")
		return c.fetch(query, retryCount-1)
	}

	content := strings.TrimPrefix(result.Choices[0].Message.Content, "```")
	content = strings.TrimPrefix(result.Choices[0].Message.Content, "json")
	content = strings.TrimSuffix(content, "```")

	c.cache.Add(query, content)

	return content
}
