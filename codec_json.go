package goconf

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"reflect"

	"github.com/adambarber/gorefuse"
)

func jsonDecodePrecheck(data SourceBytes, file FilePath) error {
	if err := refuseDuplicateJSONKeys(data, file); err != nil {
		return err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return refuseDecoder(
			gorefuse.At(gorefuse.File(file), 0,
				gorefuse.Subject("JSON syntax"),
				gorefuse.Why("the file is empty; add settings or delete it.")),
			errEmptyJSON)
	}

	return nil
}

// decodeJSONStrict decodes JSON bytes into the target, refusing unknown
// fields, trailing data, and type mismatches with the line named. Every
// refusal's frame points at the refused file.
func decodeJSONStrict(data SourceBytes, file FilePath, dst Target) error {
	typ := dst.Value.Type()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if hasExtrasInType(typ) {
		return decodeJSONStrictWithExtras(data, file, dst)
	}

	if err := jsonDecodePrecheck(data, file); err != nil {
		return err
	}

	err := json.Unmarshal(data, dst.Value.Addr().Interface(),
		json.RejectUnknownMembers(true),
		json.MatchCaseInsensitiveNames(true),
		jsontext.AllowDuplicateNames(true))
	if err != nil {
		return jsonErr(data, file, err)
	}

	return nil
}
