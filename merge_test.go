package goconf

import (
	"reflect"
	"strings"
	"testing"
)

type mergeCfg struct {
	Region  region    `json:"region"`
	Nested  nestedCfg `json:"nested"`
	Retries int       `json:"retries"`
}

// mergeCase is one layering behavior: base then overlay, merged value out.
type mergeCase struct {
	name     caseName
	wantErr  refusalNeed
	wantKeys []FieldPath
	base     SourceBytes
	overlay  SourceBytes
	want     mergeCfg
}

// mergeWant composes the expected merged value.
func mergeWant(retries int, reg region, hostCfg nestedCfg) mergeCfg {
	return mergeCfg{Nested: hostCfg, Region: reg, Retries: retries}
}

func TestMergeTwoSources(t *testing.T) {
	cases := []mergeCase{
		{
			name:     "later source overrides earlier field by field",
			base:     SourceBytes(`{"region":"eu","retries":3,"nested":{"host":"h","port":1}}`),
			overlay:  SourceBytes(`{"region":"us"}`),
			want:     mergeWant(3, "us", nestedCfg{Host: "h", Port: 1}),
			wantErr:  none,
			wantKeys: []FieldPath{fpRegion, fpRetries, fpNested, fpNestedHost, fpNestedPort},
		},
		{
			name:     "nested objects deep merge",
			base:     SourceBytes(`{"region":"eu","nested":{"host":"h","port":1}}`),
			overlay:  SourceBytes(`{"nested":{"port":2}}`),
			want:     mergeWant(0, "eu", nestedCfg{Host: "h", Port: 2}),
			wantErr:  none,
			wantKeys: []FieldPath{fpRegion, fpNested, fpNestedHost, fpNestedPort},
		},
		{
			name:     "empty overlay changes nothing",
			base:     SourceBytes(`{"region":"eu"}`),
			overlay:  SourceBytes(`{}`),
			want:     mergeWant(0, "eu", nestedCfg{Host: "", Port: 0}),
			wantErr:  none,
			wantKeys: []FieldPath{fpRegion},
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			layers := mustLayer(t, tc.base, tc.overlay)
			got := mustDecodeMerged(t, layers)

			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}

			assertPresent(t, layers, tc.wantKeys)
		})
	}
}

// assertPresent checks every wanted key is in the merged presence.
func assertPresent(t *testing.T, layers Layers, wantKeys []FieldPath) {
	t.Helper()

	for _, k := range wantKeys {
		if !layers.Present().Has(k) {
			t.Errorf("expected key %q in merged presence", k)
		}
	}
}

// mustLayer layers two sources or fails the test.
func mustLayer(t *testing.T, base, overlay SourceBytes) Layers {
	t.Helper()

	layers, err := layerSources([]SourceBytes{base, overlay})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return layers
}

// mustDecodeMerged decodes the layers or fails the test.
func mustDecodeMerged(t *testing.T, layers Layers) mergeCfg {
	t.Helper()

	var got mergeCfg

	err := decodeMerged(layers, Target{Value: reflect.ValueOf(&got).Elem()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return got
}

func TestMergeRefusesUnknownFields(t *testing.T) {
	layers, err := layerSources([]SourceBytes{SourceBytes(`{"region":"eu"}`), SourceBytes(`{"extra":1}`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got mergeCfg

	err = decodeMerged(layers, Target{Value: reflect.ValueOf(&got).Elem()})

	if err == nil || !strings.Contains(err.Error(), "extra") {
		t.Fatalf("expected unknown-field refusal naming extra, got %v", err)
	}
}

func TestMergeExplicitNullIsKept(t *testing.T) {
	layers, err := layerSources([]SourceBytes{SourceBytes(`{"region":"eu"}`), SourceBytes(`{"region":null}`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !layers.Present().Has(fpRegion) {
		t.Fatalf("explicit null is a decision and must mark the key present")
	}
}
