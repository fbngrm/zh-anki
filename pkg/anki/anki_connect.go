package anki

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// AnkiConnect API URL
const ankiConnectURL = "http://localhost:8765"

// Note struct represents the fields of an Anki note
type Note struct {
	DeckName  string            `json:"deckName"`
	ModelName string            `json:"modelName"`
	Fields    map[string]string `json:"fields"`
	Options   struct {
		AllowDuplicate bool `json:"allowDuplicate"`
	} `json:"options"`
	Tags []string `json:"tags"`
}

// Response struct represents the response from AnkiConnect
type Response struct {
	Result int    `json:"result"`
	Error  string `json:"error"`
}

// AddNoteToDeck adds a new note to the specified deck in Anki
func AddNoteToDeck(deckName, modelName string, noteFields map[string]string) (int, error) {
	// Create a new note with the provided fields
	note := Note{
		DeckName:  deckName,
		ModelName: modelName,
		Fields:    noteFields,
		Options: struct {
			AllowDuplicate bool `json:"allowDuplicate"`
		}{
			AllowDuplicate: false,
		},
		Tags: []string{},
	}

	// Prepare the request payload
	payload := struct {
		Action  string      `json:"action"`
		Version int         `json:"version"`
		Params  interface{} `json:"params"`
	}{
		Action:  "addNote",
		Version: 6,
		Params: struct {
			Note Note `json:"note"`
		}{
			Note: note,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	// Send the request to AnkiConnect
	response, err := http.Post(ankiConnectURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	if response.StatusCode == http.StatusOK {
		var responseData Response
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			return 0, err
		}

		if responseData.Error != "" {
			return 0, errors.New(responseData.Error)
		}

		return responseData.Result, nil
	}

	return 0, fmt.Errorf("failed to add note. Status code: %d", response.StatusCode)
}

// UpdateNoteInDeck updates the specified note in Anki with new fields
func UpdateNoteInDeck(noteID string, noteFields map[string]string) error {
	// Prepare the request payload
	payload := struct {
		Action  string      `json:"action"`
		Version int         `json:"version"`
		Params  interface{} `json:"params"`
	}{
		Action:  "updateNoteFields",
		Version: 6,
		Params: struct {
			Note struct {
				ID     string            `json:"id"`
				Fields map[string]string `json:"fields"`
			} `json:"note"`
		}{
			Note: struct {
				ID     string            `json:"id"`
				Fields map[string]string `json:"fields"`
			}{
				ID:     noteID,
				Fields: noteFields,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send the request to AnkiConnect
	response, err := http.Post(ankiConnectURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode == http.StatusOK {
		var responseData Response
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			return err
		}

		if responseData.Error != "" {
			return fmt.Errorf("failed to update note: %s", responseData.Error)
		}

		return nil
	}

	return fmt.Errorf("failed to update note. Status code: %d", response.StatusCode)
}
