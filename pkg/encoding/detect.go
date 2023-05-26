package encoding

import "unicode"

type RuneType int

const (
	RuneType_ASCII = iota
	RuneType_CJKUnifiedIdeograph
	RuneType_CJKRadical
	RuneType_CJKSymbolsPunctuation
	RuneType_Pinyin
)

func DetectRuneType(r rune) RuneType {
	switch {
	case unicode.Is(unicode.Han, r):
		return RuneType_CJKUnifiedIdeograph
	default:
		return RuneType_ASCII
	}
}
