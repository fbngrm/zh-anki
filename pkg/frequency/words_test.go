package frequency

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestNewWordIndex(t *testing.T) {
	// Create a temporary file for the test
	tmpfile, err := ioutil.TempFile("", "index.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test data to the temporary file
	data := []byte("word1: frequency1\nword2: frequency2\nword3: frequency3")
	if _, err := tmpfile.Write(data); err != nil {
		t.Fatalf("Failed to write data to temporary file: %v", err)
	}

	// Close the temporary file
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Call NewWordIndex with the temporary file path
	index, err := NewWordIndex(tmpfile.Name())
	if err != nil {
		t.Fatalf("NewWordIndex returned an error: %v", err)
	}

	// Verify the index path
	expectedPath := tmpfile.Name()
	if index.path != expectedPath {
		t.Errorf("Expected path: %s, but got: %s", expectedPath, index.path)
	}

	// Verify the index content
	expectedIndex := []string{"word1", "word2", "word3"}
	if !stringSlicesEqual(index.Words, expectedIndex) {
		t.Errorf("Expected index: %v, but got: %v", expectedIndex, index.Words)
	}
}

// Helper function to compare two string slices
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
func TestWordIndex_GetMostFrequent(t *testing.T) {
	wi := WordIndex{
		Words: []string{"apple", "banana", "cherry", "banana", "apple"},
	}

	expected := []string{"a", "p", "l", "e", "apple", "b", "n", "banana", "c", "h", "r", "y", "cherry"}
	result := wi.GetMostFrequent(100)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result. Expected: %v, Got: %v", expected, result)
	}
}
