package goconf

import (
	"encoding/json"

	"github.com/adambarber/gorefuse"
)

func decodeJSONObjectKey(dec *json.Decoder, data SourceBytes, file FilePath) (keySegment, error) {
	ktok, err := dec.Token()
	if err != nil {
		return "", jsonErr(data, file, err)
	}

	key, ok := ktok.(string)
	if ok {
		return keySegment(key), nil
	}

	return keySegment(""), refuseDecoder(gorefuse.At(gorefuse.File(file), gorefuse.Line(lineAtByte(data, dec.InputOffset())), gorefuse.Subject("JSON syntax"), syntaxCheck()), ErrExpectedJSONObjectKey)
}
