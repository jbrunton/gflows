package json

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func cleanKeys(dst *bytes.Buffer, src []byte) error {
	origLen := dst.Len()
	scan := newScanner()
	defer freeScanner(scan)
	//needIndent := false
	//depth := 0
	currentKey := []byte{}
	recordKey := false
	//prevParseState := -1
	fmt.Println("parseObjectKey:", parseObjectKey)
	fmt.Println("parseObjectValue:", parseObjectValue)
	fmt.Println("parseArrayValue:", parseArrayValue)
	for _, c := range src {
		scan.bytes++
		v := scan.step(scan, c)
		if v == scanError {
			break
		}

		if len(scan.parseState) > 0 {
			currentParseState := scan.parseState[len(scan.parseState)-1]
			//fmt.Println("scan.parseState:", scan.parseState)
			//fmt.Println("currentParseState:", currentParseState)
			if currentParseState == parseObjectKey {
				//fmt.Println("currentParseState == scanObjectKey")
				if len(currentKey) == 0 && c == '"' {
					recordKey = true
				} else if len(currentKey) > 0 && c == '"' {
					recordKey = false
					continue
				}

				if recordKey {
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
			//fmt.Println("scan.parseState:", scan.parseState)
			fmt.Println("key:", string(currentKey))
			dst.Write(currentKey[1:len(currentKey)])
			dst.WriteString(":")
			currentKey = []byte{}
			continue
		}
		// if v == scanBeginObject {
		// 	fmt.Println("scanBeginObject")
		// 	fmt.Println("scan.parseState:", scan.parseState)
		// 	fmt.Println("scanBeginObject:", scanBeginObject)
		// }
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
