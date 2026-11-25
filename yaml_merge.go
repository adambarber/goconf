package goconf

import "gopkg.in/yaml.v3"

// yamlMergePair is one key/value pair contributed by a << node.
type yamlMergePair struct {
	key   *yaml.Node
	value *yaml.Node
}

// yamlMergeCollector walks one merge node and records the pairs it contributes
// in document order. Explicit keys and earlier merged keys win.
type yamlMergeCollector struct {
	explicit map[keySegment]bool
	merged   map[keySegment]bool
	seen     map[*yaml.Node]bool
	pairs    []yamlMergePair
}

// yamlMergedPairs resolves one << node into the pairs it contributes. It
// mirrors the presence collector's merge walk so decode and metadata agree.
func yamlMergedPairs(node *yaml.Node, explicit map[keySegment]bool) []yamlMergePair {
	c := yamlMergeCollector{
		explicit: explicit,
		merged:   map[keySegment]bool{},
		seen:     map[*yaml.Node]bool{},
		pairs:    nil,
	}
	c.walk(node)

	return c.pairs
}

func (c *yamlMergeCollector) walk(node *yaml.Node) {
	node = yamlResolveAlias(node)
	if node == nil || c.seen[node] {
		return
	}

	c.seen[node] = true

	switch node.Kind {
	case yaml.SequenceNode:
		for _, item := range node.Content {
			c.walk(item)
		}
	case yaml.MappingNode:
		c.collectMapping(node)
	case yaml.DocumentNode, yaml.ScalarNode, yaml.AliasNode:
		return
	}
}

func (c *yamlMergeCollector) collectMapping(node *yaml.Node) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		kNode, vNode := node.Content[i], node.Content[i+1]
		key := keySegment(kNode.Value)

		if key == "<<" {
			c.walk(vNode)

			continue
		}

		if c.explicit[key] || c.merged[key] {
			continue
		}

		c.merged[key] = true
		c.pairs = append(c.pairs, yamlMergePair{key: kNode, value: vNode})
	}
}

// yamlExplicitAndMerge splits one mapping's pairs into explicit keys and the
// merge nodes that must be applied before them.
func yamlExplicitAndMerge(node *yaml.Node) (map[keySegment]bool, []*yaml.Node) {
	explicit := map[keySegment]bool{}

	var merges []*yaml.Node

	for i := 0; i+1 < len(node.Content); i += 2 {
		kNode, vNode := node.Content[i], node.Content[i+1]
		if keySegment(kNode.Value) == "<<" {
			merges = append(merges, vNode)

			continue
		}

		explicit[keySegment(kNode.Value)] = true
	}

	return explicit, merges
}
