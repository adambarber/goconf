package goconf

import (
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

func yamlNodeValue(node *yaml.Node) reflect.Value {
	if node == nil {
		return reflect.Value{}
	}

	if node.Kind == yaml.AliasNode {
		return yamlAliasValue(node)
	}

	switch node.Kind {
	case yaml.ScalarNode:
		return yamlScalarValue(node)
	case yaml.MappingNode, yaml.SequenceNode, yaml.DocumentNode, yaml.AliasNode:
		return reflect.Value{}
	default:
		return reflect.Value{}
	}
}

func yamlAliasValue(node *yaml.Node) reflect.Value {
	if node.Alias == nil {
		return reflect.Value{}
	}

	return yamlNodeValue(node.Alias)
}

func yamlScalarValue(node *yaml.Node) reflect.Value {
	switch node.Tag {
	case "!!bool":
		b, err := strconv.ParseBool(node.Value)
		if err == nil {
			return reflect.ValueOf(b)
		}
	case "!!int":
		if i, err := strconv.ParseInt(node.Value, 0, 64); err == nil {
			return reflect.ValueOf(i)
		}
	case "!!float":
		if f, err := strconv.ParseFloat(node.Value, 64); err == nil {
			return reflect.ValueOf(f)
		}
	case "!!null":
		return reflect.Value{}
	case "!!str":
		return reflect.ValueOf(node.Value)
	default:
		return yamlScalarImplicit(node)
	}

	return yamlScalarImplicit(node)
}

func yamlScalarImplicit(node *yaml.Node) reflect.Value {
	switch node.Value {
	case "~", "null", "Null", "NULL", "":
		if node.Tag == "" || node.Tag == "!" || node.Tag == "!!null" {
			return reflect.Value{}
		}
	case "true", "True", "TRUE", "false", "False", "FALSE":
		if b, err := strconv.ParseBool(node.Value); err == nil {
			return reflect.ValueOf(b)
		}
	}

	if i, err := strconv.ParseInt(node.Value, 0, 64); err == nil {
		return reflect.ValueOf(i)
	}

	if f, err := strconv.ParseFloat(node.Value, 64); err == nil {
		return reflect.ValueOf(f)
	}

	return reflect.ValueOf(node.Value)
}
