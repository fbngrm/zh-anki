package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode/utf8"
)

const outputDir = "/home/f/work/src/github.com/fbngrm/zh-anki/data/daily/"
const ankiConnectURL = "http://127.0.0.1:8765"

// AnkiRequest represents the request body for AnkiConnect.
type AnkiRequest struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

// AnkiResponse represents the response body from AnkiConnect.
type AnkiResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

// CardInfo represents detailed information about a card.
type CardInfo struct {
	NoteType string                 `json:"modelName"`
	Fields   map[string]interface{} `json:"fields"`
}

func main() {
	cardIDs := fetchDueCards()
	cards := fetchCardInfo(cardIDs)
	classifyAndStoreCards(cards[:100], "due")

	cardIDs = fetchNewCards()
	cards = fetchCardInfo(cardIDs)
	classifyAndStoreCards(cards[:10], "new")
}

// fetchNewCards retrieves up to `limit` due cards from Anki.
func fetchNewCards() []int {
	requestBody := AnkiRequest{
		Action:  "findCards",
		Version: 6,
		Params: map[string]string{
			"query": fmt.Sprintf(`deck:"chinese::zh" is:new`),
		},
	}
	return fetch(requestBody)
}

// fetchDueCards retrieves up to `limit` due cards from Anki.
func fetchDueCards() []int {
	requestBody := AnkiRequest{
		Action:  "findCards",
		Version: 6,
		Params: map[string]string{
			"query": fmt.Sprintf(`deck:"chinese::zh" is:due`),
		},
	}
	return fetch(requestBody)
}

func fetch(requestBody AnkiRequest) []int {
	response := sendRequest(requestBody)
	cardIDs, ok := response.Result.([]interface{})
	if !ok {
		fmt.Println("Unexpected result format for findCards:", response.Result)
		os.Exit(1)
	}

	var ids []int
	for _, id := range cardIDs {
		ids = append(ids, int(id.(float64)))
	}
	return ids
}

// fetchCardInfo retrieves detailed information for the given card IDs.
func fetchCardInfo(cardIDs []int) []CardInfo {
	requestBody := AnkiRequest{
		Action:  "cardsInfo",
		Version: 6,
		Params: map[string][]int{
			"cards": cardIDs,
		},
	}

	response := sendRequest(requestBody)
	cardInfoList, ok := response.Result.([]interface{})
	if !ok {
		fmt.Println("Unexpected result format for cardsInfo:", response.Result)
		os.Exit(1)
	}

	var cards []CardInfo
	for _, cardInfo := range cardInfoList {
		var card CardInfo
		cardData, _ := json.Marshal(cardInfo)
		if err := json.Unmarshal(cardData, &card); err != nil {
			fmt.Printf("unmarshal card %v: %v\n", cardData, err)
		}
		cards = append(cards, card)
	}
	return cards
}

// classifyAndStoreCards processes cards and stores them in respective files.
func classifyAndStoreCards(cards []CardInfo, prefix string) {
	outDir := path.Join(outputDir, prefix, "input")
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		fmt.Println("create dir:", err)
	}

	wordFile, _ := os.Create(path.Join(outDir, "words"))
	clozeFile, _ := os.Create(path.Join(outDir, "clozes"))
	sentenceFile, _ := os.Create(path.Join(outDir, "sentences"))
	defer wordFile.Close()
	defer clozeFile.Close()
	defer sentenceFile.Close()

	num := 0
	for _, card := range cards {
		// fmt.Println("type:", card.NoteType)
		noteType := card.NoteType
		chineseField := getFieldCaseInsensitive(card.Fields, "Chinese")
		// fmt.Println("ch:", chineseField)
		if chineseField == "" {
			continue
		}
		switch noteType {
		case "word_cedict3", "word":
			writeToFile(wordFile, chineseField)
		case "cloze":
			clozeSentenceFront := card.Fields["SentenceFront"]
			clozeSentenceFront = strings.ReplaceAll(clozeSentenceFront.(map[string]interface{})["value"].(string), "_", "("+chineseField+")")
			writeToFile(clozeFile, fmt.Sprintf("%s\t%s", chineseField, clozeSentenceFront))
		case "sentence":
			writeToFile(sentenceFile, chineseField)
		default:
			// processRemainingTypes(noteType, chineseField, wordFile, sentenceFile)
			continue
		}
		num++
		if num == 100 {
			break
		}
	}
	fmt.Println("Cards have been classified and stored in their respective files.")
}

// processRemainingTypes handles classification of cards not matching the predefined types.
func processRemainingTypes(field string, wordFile, sentenceFile *os.File) {
	runeCount := utf8.RuneCountInString(field)
	if runeCount == 1 {
		// Ignore single-rune fields
		return
	}
	if runeCount >= 2 && runeCount <= 4 && !strings.Contains(field, " ") {
		writeToFile(wordFile, field)
	} else if runeCount > 4 && strings.Contains(field, " ") {
		writeToFile(sentenceFile, field)
	}
}

// getFieldCaseInsensitive retrieves a field value by key, ignoring case.
func getFieldCaseInsensitive(fields map[string]interface{}, key string) string {
	for k, v := range fields {
		if strings.EqualFold(k, key) {
			return v.(map[string]interface{})["value"].(string)
		}
	}
	return ""
}

// writeToFile writes a line to the given file.
func writeToFile(file *os.File, line string) {
	_, err := file.WriteString(line + "\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		os.Exit(1)
	}
}

// sendRequest sends an HTTP POST request to AnkiConnect and returns the response.
func sendRequest(requestBody AnkiRequest) AnkiResponse {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		os.Exit(1)
	}

	resp, err := http.Post(ankiConnectURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request to AnkiConnect:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	var response AnkiResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		os.Exit(1)
	}

	if response.Error != "" {
		fmt.Println("AnkiConnect returned an error:", response.Error)
		os.Exit(1)
	}

	return response
}
