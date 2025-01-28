package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type FindCardsRequest struct {
	Action  string              `json:"action"`
	Version int                 `json:"version"`
	Params  map[string][]string `json:"params"`
}

type FindCardsResponse struct {
	Result map[string]interface{} `json:"result"`
	Error  string                 `json:"error"`
}

func getDeckStats(deckName string) {
	// Prepare request
	request := FindCardsRequest{
		Action:  "getDeckStats",
		Version: 6,
		Params: map[string][]string{
			"decks": {"chinese::zh"},
		},
	}

	// Convert to JSON
	payload, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error encoding request:", err)
		return
	}

	// Send request
	resp, err := http.Post("http://localhost:8765", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Decode response
	var response FindCardsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	// Handle errors
	if response.Error != "" {
		fmt.Printf("Error fetching: %s\n", response.Error)
		return
	}

	fmt.Printf("%v: \n", response.Result)
}

func main() {
	// Replace "Default" with your deck name
	deckName := "zh"
	getDeckStats(deckName)
}
