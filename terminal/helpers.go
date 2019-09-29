package terminal

import (
	"bytes"
	"strings"
)

var (
	QuoteChars   = `"'`
	BreakChars   = ` '"`
	EscapedChars = ` \`
)

func isCharQuoted(line string, index int) bool {
	return index > 0 &&
		line[index-1] == '\\' &&
		!isCharQuoted(line, index-1)

}

func LineSplit(line string) []string {
	var quoteChar rune
	var words []string
	var j int
	for i, r := range line {
		if quoteChar != 0 {
			if r == quoteChar {
				quoteChar = 0
			}
		} else if strings.IndexRune(QuoteChars, r) >= 0 {
			quoteChar = r
		}

		if strings.IndexRune(BreakChars, r) >= 0 {
			if ((quoteChar == 0 && strings.IndexRune(QuoteChars, r) < 0) || (quoteChar == r && strings.IndexRune(line[i+1:], r) < 0)) && !isCharQuoted(line, i) {
				words = append(words, line[j:i+1])
				j = i + 1
			}
		}
	}
	words = append(words, line[j:])
	return words
}

func EscapeString(s string, withQuote bool) string {
	wordBuff := &bytes.Buffer{}
	if len(s) > 0 {
		for _, c := range s[:len(s)-1] {
			if strings.IndexRune(EscapedChars, c) >= 0 && !withQuote {
				wordBuff.WriteByte('\\')
			}
			wordBuff.WriteRune(c)
		}
		wordBuff.WriteByte(s[len(s)-1])
	}
	return wordBuff.String()
}

func UnescapeString(s string) string {
	wordBuff := &bytes.Buffer{}
	len := len(s)
	for i := 0; i < len; i++ {
		if s[i] == '\\' {
			if i < len-1 && strings.IndexByte(EscapedChars, s[i+1]) >= 0 {
				i++
			}
		}
		wordBuff.WriteByte(s[i])
	}
	return wordBuff.String()
}
