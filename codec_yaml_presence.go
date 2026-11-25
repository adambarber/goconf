package goconf

import (
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

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

			return docs, decodeRefusal(data, file, errValueFree(err))
		}

		docs = append(docs, doc)
	}
}

// shapeUserTerm is the user's word for one document node's shape.
func shapeUserTerm(node *yaml.Node) userTerm {
	switch {
	case node.Kind == yaml.SequenceNode:
		return kindUserTerm(kindList)
	case node.Kind == yaml.ScalarNode:
		return kindUserTerm(kindString)
	default:
		return kindUserTerm(kindTable)
	}
}
