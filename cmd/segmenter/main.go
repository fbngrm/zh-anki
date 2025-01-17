package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// segmentChineseText uses the Stanford Segmenter to segment a Chinese string.
func segmentChineseText(input string) (string, error) {
	segmentScript := "/home/f/work/src/github.com/fbngrm/stanford-segmenter/segment.sh"

	// The model to use (e.g., "ctb" for Penn Chinese Treebank)
	segmenterModel := "pku"

	tempInputFile, err := ioutil.TempFile("", "input-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary input file: %v", err)
	}
	defer os.Remove(tempInputFile.Name()) // Ensure the temp file is removed

	// Write the input string to the temporary file
	if _, err := tempInputFile.WriteString(input); err != nil {
		return "", fmt.Errorf("failed to write to temporary input file: %v", err)
	}
	tempInputFile.Close()

	tempOutputFile, err := ioutil.TempFile("", "1output-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary output file: %v", err)
	}
	// defer os.Remove(tempOutputFile.Name()) // Ensure the temp file is removed
	tempOutputFile.Close()

	// Build the command to execute the segment.sh script
	cmd := exec.Command(segmentScript, segmenterModel, tempInputFile.Name(), "UTF-8", "0")

	// Capture both stdout and stderr for debugging
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Run the command
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running segment.sh: %v", err)
	}

	// Return the segmented output as a string
	return stdout.String(), nil
}

func main() {
	// Example input
	input := "我需要复习今天的课文。"

	// Call the function
	segmentedText, err := segmentChineseText(input)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the segmented text
	fmt.Println("Segmented Text:", segmentedText)
}
