package goconf

import (
	"errors"
	"fmt"
	"io"
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

	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("yaml: %w %d", errMultiDoc, extra.Line)
	}

	return nil
}

// errMultiDoc is the refusal for a stream of more than one YAML
// document: first-wins silently resolves what must be refused.
var errMultiDoc = errors.New("stream holds more than one document; separator at line")

// presentYAMLKeys records the dotted field paths of every key present in
// the YAML document, walking the node tree so presence is recorded no
// matter how deeply it is nested, including inside sequences of mappings.
func presentYAMLKeys(data SourceBytes) (KeySet, error) {
	keys := KeySet{paths: nil}

	docs, err := yamlDocuments(data)
	if err != nil {
		return keys, err
	}

	if len(docs) == 0 {
		return keys, nil
	}

	if len(docs) > 1 {
		return keys, fmt.Errorf("yaml: %w %d", errMultiDoc, docs[1].Line)
	}

	if len(docs[0].Content) == 0 {
		return keys, nil
	}

	root := docs[0].Content[0]
	if root.Kind != yaml.MappingNode {
		return keys, fmt.Errorf("yaml: line %d: %w", root.Line, errYAMLNotAMapping)
	}

	walkYAMLNode(&keys, "", root)

	return keys, nil
}

// yamlDocuments decodes every document of the stream as a bare node;
// the second document's Line names the separator for the refusal. The
// second parse (next to decodeYAMLStrict's) guards presence, not merges.
func yamlDocuments(data SourceBytes) ([]yaml.Node, error) {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))

	docs := []yaml.Node{}

	for {
		var doc yaml.Node
		if err := dec.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				return docs, nil
			}

			return docs, fmt.Errorf("yaml: %w", errValueFree(err))
		}

		docs = append(docs, doc)
	}
}

// walkYAMLNode records every mapping key under the dotted prefix. A
// mapping node's content alternates key node, value node.
func walkYAMLNode(keys *KeySet, prefix FieldPath, node *yaml.Node) {
	walkYAMLNodeSeen(keys, prefix, node, map[*yaml.Node]bool{})
}

// walkYAMLChildSeen descends into one value node: mappings recurse,
// sequences recurse per element, aliases resolve to their anchor's node
// with a visited set against cycles, scalars are leaves.
func walkYAMLChildSeen(keys *KeySet, prefix FieldPath, node *yaml.Node, seen map[*yaml.Node]bool) {
	if node.Kind == yaml.AliasNode {
		if seen[node] || node.Alias == nil {
			return
		}

		seen[node] = true
		walkYAMLChildSeen(keys, prefix, node.Alias, seen)

		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		walkYAMLNodeSeen(keys, prefix, node, seen)
	case yaml.SequenceNode:
		for _, item := range node.Content {
			walkYAMLChildSeen(keys, prefix, item, seen)
		}
	case yaml.DocumentNode, yaml.ScalarNode, yaml.AliasNode:
		// leaves and stream roots contribute no keys
	}
}

// walkYAMLNodeSeen is walkYAMLNode carrying the alias-visited set.
func walkYAMLNodeSeen(keys *KeySet, prefix FieldPath, node *yaml.Node, seen map[*yaml.Node]bool) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		p := under(keys, prefix, FieldPath(node.Content[i].Value))
		walkYAMLChildSeen(keys, p, node.Content[i+1], seen)
	}
}
