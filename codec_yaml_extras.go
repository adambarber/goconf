package goconf

import (
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/adambarber/gorefuse"
	"gopkg.in/yaml.v3"
)

func decodeYAMLStrictWithExtras(data SourceBytes, file FilePath, dst Target) error {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(false)

	var doc yaml.Node
	if err := dec.Decode(&doc); err != nil {
		return yamlDecodeRefusal(data, file, dst.Value.Type(), err)
	}
	// refuse multi-document streams
	var extra yaml.Node

	multiErr := dec.Decode(&extra)
	if !errors.Is(multiErr, io.EOF) {
		return multiDocRefusal(file, extra.Line)
	}
	// walk node and handle extras/validate unknowns
	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	v := dst.Value
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if err := yamlPopulateNode(root, v, file); err != nil {
		return err
	}

	return nil
}

func yamlPopulateNode(node *yaml.Node, dst reflect.Value, file FilePath) error {
	if node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}

	dst = derefAlloc(dst)

	k := dst.Kind()
	if k == reflect.Struct {
		return yamlPopulateStructNode(node, dst, file)
	}

	if k == reflect.Map {
		return yamlPopulateMapNode(node, dst, file)
	}

	if k == reflect.Slice {
		return yamlPopulateSliceNode(node, dst, file)
	}

	return yamlDecodeScalarNode(node, dst, file)
}

func yamlPopulateStructNode(node *yaml.Node, dst reflect.Value, file FilePath) error {
	if node.Kind == yaml.MappingNode {
		return yamlPopulateStruct(node, dst, file)
	}

	return typeMismatchRefusal(file, node.Line, "", kindTable, shapeUserTerm(node))
}

func yamlPopulateMapNode(node *yaml.Node, dst reflect.Value, file FilePath) error {
	switch node.Kind {
	case yaml.MappingNode:
		return yamlPopulateMap(node, dst, file)
	case yaml.DocumentNode, yaml.SequenceNode, yaml.ScalarNode, yaml.AliasNode:
		return typeMismatchRefusal(file, node.Line, "", kindTable, shapeUserTerm(node))
	default:
		return typeMismatchRefusal(file, node.Line, "", kindTable, shapeUserTerm(node))
	}
}

func yamlPopulateSliceNode(node *yaml.Node, dst reflect.Value, file FilePath) error {
	if node.Kind != yaml.SequenceNode {
		return typeMismatchRefusal(file, node.Line, "", kindList, shapeUserTerm(node))
	}

	return yamlPopulateSlice(node, dst, file)
}

func yamlDecodeScalarNode(node *yaml.Node, dst reflect.Value, file FilePath) error {
	if err := node.Decode(dst.Addr().Interface()); err != nil {
		return refuseDecoder(gorefuse.At(gorefuse.File(file), gorefuse.Line(node.Line), gorefuse.Subject("YAML value"), gorefuse.Why("the value does not fit the expected type.")), err)
	}

	return nil
}

func yamlPopulateStruct(node *yaml.Node, dst reflect.Value, file FilePath) error {
	initYAMLExtrasMap(dst)

	for i := 0; i+1 < len(node.Content); i += 2 {
		if err := yamlPopulateStructPair(dst, node.Content[i], node.Content[i+1], file); err != nil {
			return err
		}
	}

	return nil
}

func initYAMLExtrasMap(dst reflect.Value) { initExtrasMap(dst) }

func yamlPopulateStructPair(dst reflect.Value, kNode, vNode *yaml.Node, file FilePath) error {
	key := keySegment(kNode.Value)

	sf, ok := matchStructField(dst.Type(), key, yamlFieldMatches)
	if ok {
		return yamlPopulateNode(vNode, dst.FieldByIndex(sf.Index), file)
	}

	if ef, ok := extrasField(dst.Type()); ok {
		return yamlSetExtra(vNode, dst.FieldByIndex(ef.Index), key, file)
	}

	return refuseAt(ErrUnknownField, file, gorefuse.Line(kNode.Line),
		gorefuse.Subject(`"`+string(key)+`"`),
		gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
}

func yamlPopulateMap(node *yaml.Node, dst reflect.Value, file FilePath) error {
	if dst.IsNil() && dst.CanSet() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		kNode := node.Content[i]
		vNode := node.Content[i+1]
		key := kNode.Value

		elem := reflect.New(dst.Type().Elem()).Elem()
		if err := yamlPopulateNode(vNode, elem, file); err != nil {
			return err
		}

		dst.SetMapIndex(reflect.ValueOf(key).Convert(dst.Type().Key()), elem)
	}

	return nil
}

func yamlPopulateSlice(node *yaml.Node, dst reflect.Value, file FilePath) error {
	slice := reflect.MakeSlice(dst.Type(), 0, len(node.Content))
	for _, item := range node.Content {
		elem := reflect.New(dst.Type().Elem()).Elem()
		if err := yamlPopulateNode(item, elem, file); err != nil {
			return err
		}

		slice = reflect.Append(slice, elem)
	}

	dst.Set(slice)

	return nil
}

func yamlSetExtra(node *yaml.Node, extras reflect.Value, key keySegment, file FilePath) error {
	if extras.Kind() != reflect.Map {
		return refuseAt(ErrUnknownField, file, gorefuse.Line(node.Line),
			gorefuse.Subject(`"`+key+`"`),
			gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
	}

	if extras.IsNil() && extras.CanSet() {
		extras.Set(reflect.MakeMap(extras.Type()))
	}

	elem := reflect.New(extras.Type().Elem()).Elem()
	if err := yamlPopulateNode(node, elem, file); err != nil {
		return err
	}

	extras.SetMapIndex(reflect.ValueOf(string(key)).Convert(extras.Type().Key()), elem)

	return nil
}
