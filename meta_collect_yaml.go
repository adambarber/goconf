package goconf

import (
	"strconv"

	"github.com/adambarber/gorefuse"
	"gopkg.in/yaml.v3"
)

func collectYAMLMeta(src SourceBytes, file FilePath, sourceIndex int) (KeySet, []Occurrence, error) {
	c := newMetaCollector(file, sourceIndex)

	docs, err := yamlDocuments(src, file)
	if err != nil {
		return KeySet{}, nil, err
	}

	if len(docs) == 0 {
		keys, entries := c.result()

		return keys, entries, nil
	}

	if len(docs) > 1 {
		return KeySet{}, nil, multiDocRefusal(file, docs[1].Line)
	}

	if len(docs[0].Content) == 0 {
		keys, entries := c.result()

		return keys, entries, nil
	}

	root := docs[0].Content[0]
	if root.Kind != yaml.MappingNode {
		return KeySet{}, nil, refuseAt(ErrNotAMapping, file, gorefuse.Line(root.Line),
			gorefuse.Subject("the file holds "+string(shapeUserTerm(root))+", but settings must be name: value pairs"),
			gorefuse.Why(`start each setting on its own line as "name: value".`))
	}

	seen := map[*yaml.Node]bool{}
	walkYAMLMeta(c, "", "", root, seen)
	keys, entries := c.result()

	return keys, entries, nil
}

func walkYAMLMeta(c *metaCollector, prefix OccurrencePath, schema SchemaPath, node *yaml.Node, seen map[*yaml.Node]bool) {
	if node == nil {
		return
	}

	if node.Kind == yaml.AliasNode {
		walkYAMLAliasMeta(c, prefix, schema, node, seen)

		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		walkYAMLMappingMeta(c, prefix, schema, node, seen)
	case yaml.SequenceNode:
		walkYAMLSequenceMeta(c, prefix, schema, node, seen)
	case yaml.DocumentNode:
		if len(node.Content) == 1 {
			walkYAMLMeta(c, prefix, schema, node.Content[0], seen)
		}
	case yaml.ScalarNode, yaml.AliasNode:
		return
	default:
		return
	}
}

func walkYAMLAliasMeta(c *metaCollector, prefix OccurrencePath, schema SchemaPath, node *yaml.Node, seen map[*yaml.Node]bool) {
	if node.Alias == nil || seen[node] {
		return
	}

	seen[node] = true
	walkYAMLMeta(c, prefix, schema, node.Alias, seen)
}

func walkYAMLMappingMeta(c *metaCollector, prefix OccurrencePath, schema SchemaPath, node *yaml.Node, seen map[*yaml.Node]bool) {
	explicit := yamlExplicitKeys(node)
	for i := 0; i < len(node.Content)-1; i += 2 {
		yamlMetaMapPair(c, prefix, schema, node.Content[i], node.Content[i+1], explicit, seen)
	}
}

func yamlExplicitKeys(node *yaml.Node) map[OccurrenceKey]bool {
	keys := map[OccurrenceKey]bool{}

	for i := 0; i < len(node.Content)-1; i += 2 {
		k := OccurrenceKey(node.Content[i].Value)
		if k == "<<" {
			continue
		}

		keys[k] = true
	}

	return keys
}

func yamlMetaMapPair(c *metaCollector, prefix OccurrencePath, schema SchemaPath, kNode, vNode *yaml.Node, explicit map[OccurrenceKey]bool, seen map[*yaml.Node]bool) {
	key := OccurrenceKey(kNode.Value)
	if key == "<<" {
		walkYAMLMergeMeta(c, prefix, schema, vNode, explicit, seen)

		return
	}

	yamlMetaEmitPair(c, prefix, schema, kNode, vNode, seen)
}

func yamlMetaEmitPair(c *metaCollector, prefix OccurrencePath, schema SchemaPath, kNode, vNode *yaml.Node, seen map[*yaml.Node]bool) {
	if kNode.Kind != yaml.ScalarNode {
		return
	}

	key := OccurrenceKey(kNode.Value)
	if key == "" {
		return
	}

	seg := metaSeg(key)
	op := occPath(prefix, seg)
	sp := schemaPath(schema, seg)

	c.add(op, sp, key, kNode.Line, yamlNodeValue(vNode))
	walkYAMLMeta(c, op, sp, vNode, seen)
}

func walkYAMLMergeMeta(c *metaCollector, prefix OccurrencePath, schema SchemaPath, node *yaml.Node, explicit map[OccurrenceKey]bool, seen map[*yaml.Node]bool) {
	node = yamlResolveAlias(node)
	if node == nil {
		return
	}

	merged := map[OccurrenceKey]bool{}
	walkYAMLMergeMetaSeen(c, prefix, schema, node, explicit, merged, seen)
}

func walkYAMLMergeMetaSeen(c *metaCollector, prefix OccurrencePath, schema SchemaPath, node *yaml.Node, explicit, merged map[OccurrenceKey]bool, seen map[*yaml.Node]bool) {
	node = yamlResolveAlias(node)
	if node == nil {
		return
	}

	if node.Kind == yaml.SequenceNode {
		for _, item := range node.Content {
			walkYAMLMergeMetaSeen(c, prefix, schema, item, explicit, merged, seen)
		}

		return
	}

	if node.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(node.Content)-1; i += 2 {
		kNode := node.Content[i]
		vNode := node.Content[i+1]

		key := OccurrenceKey(kNode.Value)
		if explicit[key] || merged[key] {
			continue
		}

		merged[key] = true

		yamlMetaEmitPair(c, prefix, schema, kNode, vNode, seen)
	}
}

func yamlResolveAlias(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}

	return node
}

func walkYAMLSequenceMeta(c *metaCollector, prefix OccurrencePath, schema SchemaPath, node *yaml.Node, seen map[*yaml.Node]bool) {
	for i, child := range node.Content {
		elemKey := OccurrenceKey("[" + strconv.Itoa(i) + "]")
		op := occIndex(prefix, i)
		sp := schemaIndex(schema)

		c.add(op, sp, elemKey, child.Line, yamlNodeValue(child))
		walkYAMLMeta(c, op, sp, child, seen)
	}
}
