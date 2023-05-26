package encoding

import "testing"

func TestDetectRuneType(t *testing.T) {
	testCases := []struct {
		r        rune
		expected RuneType
	}{
		{r: 'A', expected: RuneType_ASCII},
		{r: '中', expected: RuneType_CJKUnifiedIdeograph},
		// {r: 'ā', expected: RuneType_Pinyin},
		{r: '!', expected: RuneType_ASCII},
	}

	for _, tc := range testCases {
		result := DetectRuneType(tc.r)
		if result != tc.expected {
			t.Errorf("Unexpected result. Rune: %c, Expected: %v, Got: %v", tc.r, tc.expected, result)
		}
	}
}
