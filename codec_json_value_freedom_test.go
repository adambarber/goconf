package goconf

import (
	"reflect"
	"strings"
	"testing"
)

func TestJSONSyntaxRefusalNeverEchoesDecoderDetail(t *testing.T) {
	data := SourceBytes(`{"tags":["` + fuzzSentinel + `","b"`)

	var got dupCfg

	err := fileDecode(decodeJSONStrict, fileNameJSON)(data, Target{Value: reflect.ValueOf(&got).Elem()})
	if err == nil {
		t.Fatal("malformed JSON was accepted")
	}

	text := err.Error()
	if strings.Contains(text, fuzzSentinel) {
		t.Fatal("refusal echoed a value")
	}

	if strings.Contains(text, "offset") {
		t.Fatal("refusal echoed a decoder byte offset")
	}
}
