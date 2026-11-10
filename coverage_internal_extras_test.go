package goconf

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeJSONArrayAnyCoverage(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`[1,"x",true,{"a":2},[3]]`))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil || tok != json.Delim('[') {
		t.Fatalf("tok=%v err=%v", tok, err)
	}

	dst := reflect.New(emptyJSONValueType()).Elem()
	if err := decodeJSONArrayAny(dec, SourceBytes(`[1,"x",true,{"a":2},[3]]`), FilePath("f.json"), dst, ""); err != nil {
		t.Fatalf("decode: %v", err)
	}

	got := dst.Elem()
	if got.Kind() != reflect.Slice || got.Len() != 5 {
		t.Fatalf("got=%v kind=%v", got.Interface(), got.Kind())
	}
}

func TestDecodeJSONObjectAnyCoverage(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"a":1,"b":"x","c":true,"d":[1],"e":{"k":"v"}}`))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		t.Fatalf("tok=%v err=%v", tok, err)
	}

	dst := reflect.New(emptyJSONValueType()).Elem()
	if err := decodeJSONObjectAny(dec, SourceBytes(`{"a":1,"b":"x","c":true,"d":[1],"e":{"k":"v"}}`), FilePath("f.json"), dst, ""); err != nil {
		t.Fatalf("decode: %v", err)
	}

	got := dst.Elem()
	if got.Kind() != reflect.Map {
		t.Fatalf("kind=%v", got.Kind())
	}
}

func TestSetJSONAnyNumCoverage(t *testing.T) {
	dst := reflect.New(emptyJSONValueType()).Elem()
	if err := setJSONAnyNum(dst, json.Number("42"), FilePath("f.json"), ""); err != nil {
		t.Fatalf("int: %v", err)
	}

	if dst.Elem().Int() != 42 {
		t.Fatalf("got=%v", dst.Interface())
	}

	dst = reflect.New(emptyJSONValueType()).Elem()
	if err := setJSONAnyNum(dst, json.Number("1.5"), FilePath("f.json"), ""); err != nil {
		t.Fatalf("float: %v", err)
	}

	if dst.Elem().Float() != 1.5 {
		t.Fatalf("got=%v", dst.Interface())
	}
}

func TestSetJSONBoolInterfaceCoverage(t *testing.T) {
	dst := reflect.New(emptyJSONValueType()).Elem()
	if err := setJSONBool(dst, true, FilePath("f.json"), ""); err != nil {
		t.Fatalf("err=%v", err)
	}

	if !dst.Elem().Bool() {
		t.Fatalf("got=%v", dst.Interface())
	}
}

func TestSetJSONScalarInterfaceCoverage(t *testing.T) {
	dst := reflect.New(emptyJSONValueType()).Elem()
	if err := setJSONScalar(dst, jsonText("hello"), FilePath("f.json"), ""); err != nil {
		t.Fatalf("err=%v", err)
	}

	if dst.Elem().String() != "hello" {
		t.Fatalf("got=%v", dst.Interface())
	}
}
