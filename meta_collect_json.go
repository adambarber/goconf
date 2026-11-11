package goconf

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"io"
	"reflect"
	"strconv"

	"github.com/adambarber/gorefuse"
)

func collectJSONMeta(src SourceBytes, file FilePath, sourceIndex int) (KeySet, []Occurrence, error) {
	c := newMetaCollector(file, sourceIndex)
	dec := jsontext.NewDecoder(bytes.NewReader(src), jsontext.AllowDuplicateNames(true))

	if err := jsonMetaConsume(dec, src, file, c, "", ""); err != nil {
		if errors.Is(err, io.EOF) {
			keys, entries := c.result()

			return keys, entries, nil
		}

		return KeySet{}, nil, err
	}

	keys, entries := c.result()

	return keys, entries, nil
}

func jsonMetaConsume(dec *jsontext.Decoder, src SourceBytes, file FilePath, c *metaCollector, prefix OccurrencePath, schema SchemaPath) error {
	switch dec.PeekKind() {
	case '{':
		return jsonMetaObject(dec, src, file, c, prefix, schema)
	case '[':
		return jsonMetaArray(dec, src, file, c, prefix, schema)
	case jsontext.KindNull, jsontext.KindFalse, jsontext.KindTrue, jsontext.KindString, jsontext.KindNumber, '-':
		_, err := dec.ReadValue()

		return jsonMetaErr(err, src, file)
	case jsontext.KindInvalid, jsontext.KindEndObject, jsontext.KindEndArray:
		_, err := dec.ReadToken()

		return jsonMetaErr(err, src, file)
	default:
		_, err := dec.ReadToken()

		return jsonMetaErr(err, src, file)
	}
}

func jsonMetaErr(err error, src SourceBytes, file FilePath) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, io.EOF) {
		return io.EOF
	}

	return jsonErr(src, file, err)
}

func jsonMetaObject(dec *jsontext.Decoder, src SourceBytes, file FilePath, c *metaCollector, prefix OccurrencePath, schema SchemaPath) error {
	if err := jsonMetaReadOpen(dec, src, file); err != nil {
		return err
	}

	for dec.PeekKind() != '}' {
		if err := jsonMetaObjectMember(dec, src, file, c, prefix, schema); err != nil {
			return err
		}
	}

	return jsonMetaReadClose(dec, src, file)
}

func jsonMetaReadOpen(dec *jsontext.Decoder, src SourceBytes, file FilePath) error {
	_, err := dec.ReadToken()

	return jsonMetaErr(err, src, file)
}

func jsonMetaReadClose(dec *jsontext.Decoder, src SourceBytes, file FilePath) error {
	_, err := dec.ReadToken()
	if err != nil {
		return jsonMetaErr(err, src, file)
	}

	return nil
}

func jsonMetaObjectMember(dec *jsontext.Decoder, src SourceBytes, file FilePath, c *metaCollector, prefix OccurrencePath, schema SchemaPath) error {
	keyTok, err := dec.ReadToken()
	if err != nil {
		return jsonErr(src, file, err)
	}

	if keyTok.Kind() != '"' {
		return refuseDecoder(gorefuse.At(gorefuse.File(file), gorefuse.Line(lineAtByte(src, dec.InputOffset())),
			gorefuse.Subject("JSON syntax"), syntaxCheck()), ErrExpectedJSONObjectKey)
	}

	key := keyTok.String()
	line := lineAtByte(src, dec.InputOffset())
	seg := metaSeg(key)

	return jsonMetaEmit(dec, src, file, c, occPath(prefix, seg), schemaPath(schema, seg), OccurrenceKey(key), line)
}

func jsonMetaArray(dec *jsontext.Decoder, src SourceBytes, file FilePath, c *metaCollector, prefix OccurrencePath, schema SchemaPath) error {
	if _, err := dec.ReadToken(); err != nil {
		return jsonErr(src, file, err)
	}

	for i := 0; dec.PeekKind() != ']'; i++ {
		if err := jsonMetaArrayElement(dec, src, file, c, prefix, schema, i); err != nil {
			return err
		}
	}

	_, err := dec.ReadToken()

	return jsonMetaErr(err, src, file)
}

func jsonMetaArrayElement(dec *jsontext.Decoder, src SourceBytes, file FilePath, c *metaCollector, prefix OccurrencePath, schema SchemaPath, i int) error {
	line := lineAtByte(src, dec.InputOffset())
	elemKey := OccurrenceKey("[" + strconv.Itoa(i) + "]")

	return jsonMetaEmit(dec, src, file, c, occIndex(prefix, i), schemaIndex(schema), elemKey, line)
}

func jsonMetaEmit(dec *jsontext.Decoder, src SourceBytes, file FilePath, c *metaCollector, op OccurrencePath, sp SchemaPath, key OccurrenceKey, line int) error {
	val, isContainer, err := jsonMetaPeekOrReadScalar(dec, src, file)
	if err != nil {
		return err
	}

	c.add(op, sp, key, line, val)

	if isContainer {
		return jsonMetaConsume(dec, src, file, c, op, sp)
	}

	return nil
}

func jsonMetaPeekOrReadScalar(dec *jsontext.Decoder, src SourceBytes, file FilePath) (reflect.Value, bool, error) {
	switch dec.PeekKind() {
	case '{', '[':
		return reflect.Value{}, true, nil
	case jsontext.KindInvalid, jsontext.KindEndObject, jsontext.KindEndArray,
		jsontext.KindNull, jsontext.KindFalse, jsontext.KindTrue, jsontext.KindString, jsontext.KindNumber, '-':
		return jsonMetaReadScalar(dec, src, file)
	default:
		return jsonMetaReadScalar(dec, src, file)
	}
}

func jsonMetaReadScalar(dec *jsontext.Decoder, src SourceBytes, file FilePath) (reflect.Value, bool, error) {
	raw, err := dec.ReadValue()
	if err != nil {
		return reflect.Value{}, false, jsonErr(src, file, err)
	}

	return decodeJSONRawValue(raw), false, nil
}
