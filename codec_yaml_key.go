package goconf

import (
	"reflect"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// yamlKeyAtLine re-parses the document and names the first mapping key
// whose value sits on the given line; empty when none does.
func yamlKeyAtLine(data SourceBytes, line int) ObjectKey {
	docs, err := yamlDocuments(data, FilePath(""))
	if err != nil || len(docs) == 0 {
		return ""
	}

	return yamlKeyAtLineIn(&docs[0], line)
}

// yamlIssues is one TypeError's terror lines under classification.
type yamlIssues []string

// yamlKeyAtLineIn walks one document node for a key whose value node
// sits on the given line.
func yamlKeyAtLineIn(doc *yaml.Node, line int) ObjectKey {
	found, _ := yamlKeyVisit(doc, line)

	return found
}

// yamlKeyVisit visits one node in the key search: it returns the found
// key and whether the search is done, so recursion needs no outer state.
func yamlKeyVisit(node *yaml.Node, line int) (ObjectKey, bool) {
	if node == nil {
		return "", false
	}

	if node.Kind == yaml.MappingNode {
		key, hit := mappingKeyAtLine(node, line)
		if hit {
			return key, true
		}
	}

	for _, child := range node.Content {
		key, hit := yamlKeyVisit(child, line)
		if hit {
			return key, true
		}
	}

	return "", false
}

// mappingKeyAtLine names the first key of one mapping whose value sits
// on the given line.
func mappingKeyAtLine(node *yaml.Node, line int) (ObjectKey, bool) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i+1].Line == line {
			return ObjectKey(node.Content[i].Value), true
		}
	}

	return "", false
}

// yamlWantKind maps the target field the key decodes into to the shared
// value kind the refusal names: the kind comes from reflect, so defined
// types (type port int) classify as their base kind.
func yamlWantKind(target reflect.Type, key ObjectKey) valueKind {
	if key == "" {
		return kindOf(target)
	}

	t := derefType(&target)
	if t == nil || t.Kind() != reflect.Struct {
		return kindOf(t)
	}

	f, ok := fieldMatching(t, keySegment(string(key)))
	if !ok {
		return kindTable
	}

	return kindOf(f.Type)
}

// derefType strips pointer indirection from one target type.
func derefType(t *reflect.Type) reflect.Type {
	for *t != nil && (*t).Kind() == reflect.Pointer {
		elem := (*t).Elem()

		t = &elem
	}

	return *t
}

// kindOf maps one reflect type's kind to the shared value kind.
func kindOf(t reflect.Type) valueKind {
	t = derefType(&t)
	if t == nil {
		return kindTable
	}

	switch {
	case t.Kind() == reflect.String:
		return kindString
	case isIntKind(t):
		return kindInt
	case t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64:
		return kindFloat
	case t.Kind() == reflect.Bool:
		return kindBool
	case t.Kind() == reflect.Slice || t.Kind() == reflect.Array:
		return kindList
	default:
		return kindTable
	}
}

// isIntKind reports whether one reflect type's kind is integral.
func isIntKind(t reflect.Type) bool {
	return slices.Contains([]reflect.Kind{
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
	}, t.Kind())
}

// yamlFoundTerm maps one TypeError's echoed tag ("!!str", "!!seq") to
// the user's term; empty when no issue carries a tag.
func yamlFoundTerm(issues yamlIssues) userTerm {
	for _, e := range issues {
		_, tag, ok := strings.Cut(e, yamlTagHead)
		if !ok {
			continue
		}

		return kindUserTerm(yamlTagKind(kindWord(tag)))
	}

	return ""
}

// yamlTagHead opens yaml's echoed tag span in a terror line.
const yamlTagHead = "cannot unmarshal "

// yamlTagKind maps one echoed yaml tag to the shared value kind.
func yamlTagKind(tag kindWord) valueKind {
	switch {
	case strings.HasPrefix(string(tag), "!!str"):
		return kindString
	case strings.HasPrefix(string(tag), "!!int"):
		return kindInt
	case strings.HasPrefix(string(tag), "!!float"):
		return kindFloat
	case strings.HasPrefix(string(tag), "!!bool"):
		return kindBool
	case strings.HasPrefix(string(tag), "!!seq"):
		return kindList
	case strings.HasPrefix(string(tag), "!!map"):
		return kindTable
	default:
		return kindTable
	}
}
