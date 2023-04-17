package openai

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

const systemMessage = `Add pinyin to the following sentences written in simplified Chinese. Format the result into a JSON object. The original sentence should be stored in a field called chinese, the English translation should be stored in a field called english and the pinyin should be stored in a field called piniyin. Also split each sentence into words and add JSON array with those words in a field called words. Each word should be a JSON object, the original Chinese word is stored in a field called ch, the English translation is stored in a field called en and the pinyin is stored in a field called pi.`

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
}

func NewClient(apiKey string) *Client {
	return &Client{
		endpoint: "https://api.openai.com/v1/chat/completions",
		apiKey:   apiKey,
		model:    "gpt-3.5-turbo",
	}
}

// dialog := "现在几点了？现在已经十二点了。我们去吃午饭吧？我还有一点儿工作，要㩐一下。没关系，我等你。我们今天吃什么?你想吃什么？前面开了一个新饭店。我们去那个新饭店怎么样？好，新饭店在哪儿？新饭店就在商场旁边。"

func (c *Client) Decompose(dialog string) Decomposition {
	messages := []Message{
		{
			Role:    "system",
			Content: systemMessage,
		},
		{
			Role:    "user",
			Content: dialog,
		},
	}

	payload := Request{
		Model:    c.model,
		Messages: messages,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshalling JSON payload: %v", err)
	}

	// Set up the request object
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	// var result map[string]interface{}
	var result Response
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatalf("Error decoding JSON response: %v", err)
	}

	content := strings.TrimPrefix(result.Choices[0].Message.Content, "```")
	content = strings.TrimPrefix(result.Choices[0].Message.Content, "json")
	content = strings.TrimSuffix(content, "```")

	if strings.Contains(content, "\"sentences\": [") {
		var decomp Decomposition
		err = json.Unmarshal([]byte(content), &decomp)
		if err != nil {
			log.Printf("Error parsing JSON for input %s: %v", content, err)
		}
		return decomp
	}

	var sentences []Sentence
	err = json.Unmarshal([]byte(content), &sentences)
	if err != nil {
		log.Printf("Error parsing JSON for input %s: %v", result.Choices[0].Message.Content, err)
	}
	return Decomposition{
		Sentences: sentences,
	}
}
