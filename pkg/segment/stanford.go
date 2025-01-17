package segment

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type Segmenter struct {
	// The model to use (e.g., "ctb" for Penn Chinese Treebank)
	Model string
	Cmd   string
}

// SegmentChineseText uses the Stanford Segmenter to segment a Chinese string.
func (s *Segmenter) SegmentChinese(input string) (string, error) {
	tempInputFile, err := ioutil.TempFile("", "input-*.txt")
	if err != nil {
		return "", fmt.Errorf("create temporary input file for stanford-segmenter: %v", err)
	}
	defer os.Remove(tempInputFile.Name())

	if _, err := tempInputFile.WriteString(input); err != nil {
		return "", fmt.Errorf("failed to write to temporary input file: %v", err)
	}
	tempInputFile.Close()

	cmd := exec.Command(s.Cmd, s.Model, tempInputFile.Name(), "UTF-8", "0")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running segment.sh: %v", err)
	}

	return stdout.String(), nil
}
