package jsonnet

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
	processKey := false
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
					// parseObjectKey is set from the opening `{` of an object until the `:` after the key, including whitespace.
					// But we only want to examine the characters between the `"` chars.
					if len(currentKey) == 0 {
						// We just encountered the first `"`, so start processing the key
						processKey = true
					} else {
						// We encountered the closing `"`, so stop processing it
						processKey = false
					}
					continue
				} else if processKey {
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

func Marshal(v interface{}) ([]byte, error) {
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
