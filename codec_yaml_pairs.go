package goconf

import (
	"reflect"

	"github.com/adambarber/gorefuse"
	"gopkg.in/yaml.v3"
)

// yamlPairWriter applies one key/value pair to a destination. Struct and map
// destinations share the merge/explicit ordering through yamlPopulatePairs.
type yamlPairWriter func(reflect.Value, *yaml.Node, *yaml.Node, FilePath, FieldPath) error

// yamlPopulatePairs applies merged pairs first, then explicit pairs, so an
// explicit key always wins and a << key is never captured as data.
func yamlPopulatePairs(dst reflect.Value, node *yaml.Node, file FilePath, path FieldPath, write yamlPairWriter) error {
	explicit, merges := yamlExplicitAndMerge(node)
	if err := yamlApplyMergedPairs(dst, merges, explicit, file, path, write); err != nil {
		return err
	}

	return yamlApplyExplicitPairs(dst, node, file, path, write)
}

func yamlApplyMergedPairs(dst reflect.Value, merges []*yaml.Node, explicit map[keySegment]bool, file FilePath, path FieldPath, write yamlPairWriter) error {
	for _, merge := range merges {
		for _, pair := range yamlMergedPairs(merge, explicit) {
			if err := write(dst, pair.key, pair.value, file, path); err != nil {
				return err
			}
		}
	}

	return nil
}

func yamlApplyExplicitPairs(dst reflect.Value, node *yaml.Node, file FilePath, path FieldPath, write yamlPairWriter) error {
	seen := map[ObjectKey]int{}

	for i := 0; i+1 < len(node.Content); i += 2 {
		kNode := node.Content[i]
		if keySegment(kNode.Value) == "<<" {
			continue
		}

		key := ObjectKey(kNode.Value)

		if first, dup := seen[key]; dup {
			return yamlDuplicateKeyRefusal(file, key, first)
		}

		seen[key] = kNode.Line

		if err := write(dst, kNode, node.Content[i+1], file, path); err != nil {
			return err
		}
	}

	return nil
}

func yamlPopulateStruct(node *yaml.Node, dst reflect.Value, file FilePath, path FieldPath) error {
	initYAMLExtrasMap(dst)

	return yamlPopulatePairs(dst, node, file, path, yamlPopulateStructPair)
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

	return yamlPopulatePairs(dst, node, file, path, yamlPopulateMapPair)
}

func yamlPopulateMapPair(dst reflect.Value, kNode, vNode *yaml.Node, file FilePath, path FieldPath) error {
	key := ObjectKey(kNode.Value)

	elem := reflect.New(dst.Type().Elem()).Elem()
	if err := yamlPopulateNode(vNode, elem, file, childPath(path, FieldPath(key))); err != nil {
		return err
	}

	dst.SetMapIndex(reflect.ValueOf(string(key)).Convert(dst.Type().Key()), elem)

	return nil
}
