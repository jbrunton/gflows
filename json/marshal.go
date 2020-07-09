package json

import (
	"bytes"
	"encoding/json"
	"regexp"
)

var keywords []string
var idRegex *regexp.Regexp

func isIdentifier(token []byte) bool {
	if !idRegex.Match(token) {
		return false
	}

	for _, keyword := range keywords {
		if string(token) == keyword {
			return false
		}
	}

	return true
}

func cleanKeys(dst *bytes.Buffer, src []byte) error {
	origLen := dst.Len()
	scan := newScanner()
	defer freeScanner(scan)

	currentKey := []byte{}
	recordKey := false
	for _, c := range src {
		scan.bytes++
		v := scan.step(scan, c)
		if v == scanError {
			break
		}

		if len(scan.parseState) > 0 {
			currentParseState := scan.parseState[len(scan.parseState)-1]
			if currentParseState == parseObjectKey {
				if c == '"' {
					if len(currentKey) == 0 {
						recordKey = true
					} else {
						recordKey = false
					}
					continue
				} else if recordKey {
					currentKey = append(currentKey, c)
					continue
				}
			}
		}
		// Emit semantically uninteresting bytes
		// (in particular, punctuation in strings) unmodified.
		if v == scanContinue {
			dst.WriteByte(c)
			continue
		}

		if v == scanObjectKey {
			if isIdentifier(currentKey) {
				dst.Write(currentKey)
			} else {
				dst.WriteByte('"')
				dst.Write(currentKey)
				dst.WriteByte('"')
			}
			dst.WriteString(":")
			currentKey = []byte{}
			continue
		}

		dst.WriteByte(c)
	}
	if scan.eof() == scanError {
		dst.Truncate(origLen)
		return scan.err
	}
	return nil
}

func MarshalJson(v interface{}) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = cleanKeys(&buf, b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func init() {
	keywords = []string{"assert", "else", "error", "false", "for", "function", "if", "import",
		"importstr", "in", "local", "null", "tailstrict", "then", "self", "super", "true"}
	idRegex = regexp.MustCompile("^[_a-zA-Z][_a-zA-Z0-9]*$")
}
