package goconf

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// errYAMLNotAMapping is the refusal for a document whose root is not an
// object; layered defaults need a mapping to walk.
var errYAMLNotAMapping = errors.New("document root is not a mapping")

// decodeYAMLStrict decodes YAML bytes into the target, refusing unknown
// fields and type mismatches. Explicit null is accepted and decodes to the
// field's zero value: it is a decision, not an absence.
func decodeYAMLStrict(data SourceBytes, dst Target) error {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)

	if err := dec.Decode(dst.Value.Addr().Interface()); err != nil {
		return fmt.Errorf("yaml: %w", errValueFree(err))
	}

	return nil
}

// presentYAMLKeys records the dotted field paths of every key present in
// the YAML document, walking the node tree so presence is recorded no
// matter how deeply it is nested, including inside sequences of mappings.
func presentYAMLKeys(data SourceBytes) (KeySet, error) {
	keys := KeySet{paths: nil}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return keys, fmt.Errorf("yaml: %w", errValueFree(err))
	}

	if len(root.Content) == 0 {
		return keys, nil
	}

	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return keys, fmt.Errorf("yaml: %w", errYAMLNotAMapping)
	}

	walkYAMLNode(&keys, "", doc)

	return keys, nil
}

// walkYAMLNode records every mapping key under the dotted prefix. A
// mapping node's content alternates key node, value node.
func walkYAMLNode(keys *KeySet, prefix FieldPath, node *yaml.Node) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		p := under(keys, prefix, FieldPath(node.Content[i].Value))
		walkYAMLChild(keys, p, node.Content[i+1])
	}
}

// walkYAMLChild descends into one value node: mappings recurse, sequences
// recurse per element, scalars are leaves.
func walkYAMLChild(keys *KeySet, prefix FieldPath, node *yaml.Node) {
	switch node.Kind {
	case yaml.MappingNode:
		walkYAMLNode(keys, prefix, node)
	case yaml.SequenceNode:
		for _, item := range node.Content {
			walkYAMLChild(keys, prefix, item)
		}
	case yaml.DocumentNode, yaml.ScalarNode, yaml.AliasNode:
		// leaves and stream roots contribute no keys
	}
}
