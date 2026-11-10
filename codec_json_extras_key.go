package goconf

import (
	"encoding/json/jsontext"

	"github.com/adambarber/gorefuse"
)

func decodeJSONObjectKey(dec *jsontext.Decoder, data SourceBytes, file FilePath) (keySegment, error) {
	ktok, err := dec.ReadToken()
	if err != nil {
		return "", jsonErr(data, file, err)
	}

	if ktok.Kind() == '"' {
		return keySegment(ktok.String()), nil
	}

	return keySegment(""), refuseDecoder(gorefuse.At(gorefuse.File(file), gorefuse.Line(lineAtByte(data, dec.InputOffset())), gorefuse.Subject("JSON syntax"), syntaxCheck()), ErrExpectedJSONObjectKey)
}
