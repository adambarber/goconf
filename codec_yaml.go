package goconf

import (
	"errors"
	"io"
	"strings"

	"github.com/adambarber/gorefuse"
	"gopkg.in/yaml.v3"
)

// decodeYAMLStrict decodes YAML bytes into the target, refusing unknown
// fields and type mismatches. Explicit null is accepted and decodes to the
// field's zero value: it is a decision, not an absence.
func decodeYAMLStrict(data SourceBytes, file FilePath, dst Target) error {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)

	if err := dec.Decode(dst.Value.Addr().Interface()); err != nil {
		return decodeRefusal(file, errValueFree(err))
	}

	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return multiDocRefusal(file, extra.Line)
	}

	return nil
}

// decodeRefusal rewrites one decoder error as the ErrDecoderRefused
// carrier: the stripped detail survives as the cause tail.
func decodeRefusal(file FilePath, err error) error {
	return refuseCause(
		gorefuse.At(gorefuse.File(file), 0,
			gorefuse.Subject("YAML decode"),
			gorefuse.Why("fix the error the YAML decoder reported.")), err)
}

// multiDocRefusal refuses a stream holding more than one document, naming
// the separator's line.
func multiDocRefusal(file FilePath, line int) error {
	return refuseAt(ErrMultiDoc, file, gorefuse.Line(line),
		gorefuse.Subject("second document separator"),
		gorefuse.Why("first-wins would silently resolve the stream; hold one document per file."))
}

// presentYAMLKeys records the dotted field paths of every key present in
// the YAML document, walking the node tree so presence is recorded no
// matter how deeply it is nested, including inside sequences of mappings.
func presentYAMLKeys(data SourceBytes, file FilePath) (KeySet, error) {
	keys := KeySet{paths: nil}

	docs, err := yamlDocuments(data, file)
	if err != nil {
		return keys, err
	}

	if len(docs) == 0 {
		return keys, nil
	}

	if len(docs) > 1 {
		return keys, multiDocRefusal(file, docs[1].Line)
	}

	if len(docs[0].Content) == 0 {
		return keys, nil
	}

	root := docs[0].Content[0]
	if root.Kind != yaml.MappingNode {
		return keys, refuseAt(ErrNotAMapping, file, gorefuse.Line(root.Line),
			gorefuse.Subject("document root"),
			gorefuse.Why("layered defaults need a mapping to walk; shape the document as a mapping."))
	}

	walkYAMLNode(&keys, "", root)

	return keys, nil
}

// yamlDocuments decodes every document of the stream as a bare node;
// the second document's Line names the separator for the refusal. The
// second parse (next to decodeYAMLStrict's) guards presence, not merges.
func yamlDocuments(data SourceBytes, file FilePath) ([]yaml.Node, error) {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))

	docs := []yaml.Node{}

	for {
		var doc yaml.Node
		if err := dec.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				return docs, nil
			}

			return docs, decodeRefusal(file, errValueFree(err))
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
