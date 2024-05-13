package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Check if for each word in a file a card exists in Anki collection.

const ankiConnectURL = "http://localhost:8765"

type AnkiRequest struct {
	Action  string `json:"action"`
	Version int    `json:"version"`
	Params  Params `json:"params"`
}

type Params struct {
	Query string `json:"query"`
}

func main() {
	// Read the list of Chinese words from a file
	words, err := readChineseWordsFromFile("chinese_words.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	l := len(words)

	missing := make([]string, 0)
	// Check for each word if a flashcard exists in Anki
	for i := 0; i < len(words); {
		word := words[i]
		if noteExistsInAnki(word) {
			i++
		} else {
			// Remove the word from the list if no note is found
			words = append(words[:i], words[i+1:]...)
			missing = append(missing, word)
		}
	}

	fmt.Printf("removed %d words\n", l-len(words))

	// Write the updated list back to the file
	err = writeChineseWordsToFile("chinese_words.txt", words)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	err = writeChineseWordsToFile("chinese_words_missing.txt", missing)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("Process completed successfully.")
}

func readChineseWordsFromFile(filename string) ([]string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	words := strings.Fields(string(content))
	return words, nil
}

func writeChineseWordsToFile(filename string, words []string) error {
	content := strings.Join(words, ": {}\n")
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	return err
}

func noteExistsInAnki(word string) bool {
	time.Sleep(50 * time.Millisecond)
	query := fmt.Sprintf("Chinese:*>%s<*", word)
	request := AnkiRequest{
		Action:  "findNotes",
		Version: 6,
		Params: Params{
			Query: query,
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return false
	}

	resp, err := http.Post(ankiConnectURL, "application/json", strings.NewReader(string(requestJSON)))
	if err != nil {
		fmt.Println("Error sending request to AnkiConnect:", err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return false
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error decoding response JSON:", err)
		return false
	}

	if response["error"] != nil {
		log.Fatal(response["error"].(string))
	}

	var exists bool
	if response["result"] != nil {
		exists = len(response["result"].([]interface{})) > 0
	}
	if exists {
		fmt.Printf("%s exists\n", word)
		return true
	}
	fmt.Printf("%s does not exist\n", word)
	return false
}
