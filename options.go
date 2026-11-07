package goconf

import (
	"errors"
	"fmt"
	"reflect"
)

// errUnmarshalable is the refusal for a default JSON cannot represent.
var errUnmarshalable = errors.New("default cannot marshal")

// Name is the caller-declared name of a config: files are searched as
// NAME.EXT in each directory.
type Name = configName

// Spec declares one config load: the config name and the required format.
type Spec struct {
	Format Format
	Name   Name
}

// Option customizes one Load call.

// Option customizes one Load call.
type Option interface {
	apply(s *settings)
}

// settings is the mutable configuration of one Load call.
type settings struct {
	search    []DirPath
	globalDir DirPath
	defaults  []layeredDefault
}

// layeredDefault is one defaults layer whose filled fields are exactly
// the fields absent from every source. Unexported: it is built only by
// WithDefaults and carries no constructable public surface.
type layeredDefault struct {
	value reflect.Value
}

// searchPathsOption overrides the default search path list.
type searchPathsOption struct {
	paths []DirPath
}

func (o searchPathsOption) apply(s *settings) {
	s.search = o.paths
}

// globalDirOption overrides the user config directory for one load.
type globalDirOption struct {
	dir DirPath
}

func (o globalDirOption) apply(s *settings) {
	s.globalDir = o.dir
}

// defaultsOption appends one defaults layer.
type defaultsOption struct {
	value reflect.Value
}

func (o defaultsOption) apply(s *settings) {
	s.defaults = append(s.defaults, layeredDefault{value: o.value})
}

// WithSearchPaths replaces the default discovery paths (user config
// directory then the working directory) with the given list.
func WithSearchPaths(paths ...DirPath) *searchPathsOption {
	return &searchPathsOption{paths: paths}
}

// WithGlobalDir replaces the user config directory the default discovery
// starts from; the working directory still follows as the local layer.
func WithGlobalDir(dir DirPath) *globalDirOption {
	return &globalDirOption{dir: dir}
}

// WithDefaults appends one defaults layer. The value must be marshalable
// by the spec's format and struct-shaped: a broken default PANICS here —
// a Must-style contract — and is never silently stored.
func WithDefaults[T any](value T) *defaultsOption {
	v := reflect.ValueOf(value)
	if err := marshalable(v); err != nil {
		panic(err)
	}

	if err := checkDefaultShape(v); err != nil {
		panic(err)
	}

	return &defaultsOption{value: v}
}

// errBadDefaultShape is the refusal for a default value that is not a
// struct: only a struct can fill named target fields.
var errBadDefaultShape = errors.New("WithDefaults: default value must be a struct whose fields fill the target's fields, got")

// checkDefaultShape refuses a default value that is not struct-shaped.
func checkDefaultShape(v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%w %s", errBadDefaultShape, v.Kind())
	}

	return nil
}

// marshalable refuses a default whose value JSON cannot represent: a
// channel or func field would be silently dropped at decode time.
func marshalable(v reflect.Value) error {
	return walkMarshalable(v, "")
}

// walkMarshalable refuses channel and func fields anywhere in the value.
func walkMarshalable(v reflect.Value, path FieldPath) error {
	if v.Kind() == reflect.Chan || v.Kind() == reflect.Func {
		return fmt.Errorf("WithDefaults: %w", errUnmarshalable)
	}

	if v.Kind() == reflect.Struct {
		return walkStruct(v, path)
	}
	// every other kind marshals faithfully

	return nil
}

// walkStruct walks each field of a struct value.
func walkStruct(v reflect.Value, prefix FieldPath) error {
	t := v.Type()
	for i := range t.NumField() {
		p := childPath(prefix, FieldPath(t.Field(i).Name))
		if err := walkMarshalable(v.Field(i), p); err != nil {
			return err
		}
	}

	return nil
}
