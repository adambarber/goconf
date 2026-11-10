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

	v := derefValue(dst.Value)

	if err := yamlPopulateNode(root, v, file, ""); err != nil {
		return err
	}

	return nil
}

func yamlPopulateNode(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	if node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}

	dst = derefAlloc(dst)

	k := dst.Kind()
	if k == reflect.Struct {
		return yamlPopulateStructNode(node, dst, file, path)
	}

	if k == reflect.Map {
		return yamlPopulateMapNode(node, dst, file, path)
	}

	if k == reflect.Slice {
		return yamlPopulateSliceNode(node, dst, file, path)
	}

	return yamlDecodeScalarNode(node, dst, file, path)
}

func yamlPopulateStructNode(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	if node.Kind == yaml.MappingNode {
		return yamlPopulateStruct(node, dst, file, path)
	}

	if path == "" {
		return yamlExtrasRootNotMapping(file, node)
	}

	return typeMismatchRefusal(file, node.Line, ObjectKey(path), kindTable, shapeUserTerm(node))
}

func yamlPopulateMapNode(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	switch node.Kind {
	case yaml.MappingNode:
		return yamlPopulateMap(node, dst, file, path)
	case yaml.DocumentNode, yaml.SequenceNode, yaml.ScalarNode, yaml.AliasNode:
		if path == "" {
			return yamlExtrasRootNotMapping(file, node)
		}

		return typeMismatchRefusal(file, node.Line, ObjectKey(path), kindTable, shapeUserTerm(node))
	default:
		if path == "" {
			return yamlExtrasRootNotMapping(file, node)
		}

		return typeMismatchRefusal(file, node.Line, ObjectKey(path), kindTable, shapeUserTerm(node))
	}
}

func yamlPopulateSliceNode(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	if node.Kind != yaml.SequenceNode {
		return typeMismatchRefusal(file, node.Line, ObjectKey(path), kindList, shapeUserTerm(node))
	}

	return yamlPopulateSlice(node, dst, file, path)
}

func yamlDecodeScalarNode(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	if err := node.Decode(dst.Addr().Interface()); err != nil {
		return yamlNodeDecodeRefusal(file, node, dst, path, err)
	}

	return nil
}

func yamlPopulateStruct(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	initYAMLExtrasMap(dst)

	seen := map[ObjectKey]int{}

	for i := 0; i+1 < len(node.Content); i += 2 {
		kNode := node.Content[i]
		key := ObjectKey(kNode.Value)

		if first, dup := seen[key]; dup {
			return yamlDuplicateKeyRefusal(file, key, first)
		}

		seen[key] = kNode.Line

		if err := yamlPopulateStructPair(dst, kNode, node.Content[i+1], file, path); err != nil {
			return err
		}
	}

	return nil
}

func initYAMLExtrasMap(dst reflect.Value) { initExtrasMap(dst) }

func yamlPopulateStructPair(dst reflect.Value, kNode, vNode *yaml.Node, file FilePath, path FieldPath) error {
	key := keySegment(kNode.Value)

	sf, ok := matchStructField(dst.Type(), key, yamlFieldMatches)
	if ok {
		return yamlPopulateNode(vNode, dst.FieldByIndex(sf.Index), file, childPath(path, tagPathOf(sf)))
	}

	if ef, ok := extrasField(dst.Type()); ok {
		return yamlSetExtra(vNode, dst.FieldByIndex(ef.Index), key, file, childPath(path, FieldPath(key)))
	}

	return refuseAt(ErrUnknownField, file, gorefuse.Line(kNode.Line),
		gorefuse.Subject(`"`+string(key)+`"`),
		gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
}

func yamlPopulateMap(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	if dst.IsNil() && dst.CanSet() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	seen := map[ObjectKey]int{}

	for i := 0; i+1 < len(node.Content); i += 2 {
		kNode := node.Content[i]
		vNode := node.Content[i+1]
		key := ObjectKey(kNode.Value)

		if first, dup := seen[key]; dup {
			return yamlDuplicateKeyRefusal(file, key, first)
		}

		seen[key] = kNode.Line

		elem := reflect.New(dst.Type().Elem()).Elem()
		if err := yamlPopulateNode(vNode, elem, file, childPath(path, FieldPath(key))); err != nil {
			return err
		}

		dst.SetMapIndex(reflect.ValueOf(string(key)).Convert(dst.Type().Key()), elem)
	}

	return nil
}

func yamlPopulateSlice(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	slice := reflect.MakeSlice(dst.Type(), 0, len(node.Content))
	for _, item := range node.Content {
		elem := reflect.New(dst.Type().Elem()).Elem()
		if err := yamlPopulateNode(item, elem, file, path); err != nil {
			return err
		}

		slice = reflect.Append(slice, elem)
	}

	dst.Set(slice)

	return nil
}
