package goconf

import (
	"reflect"

	"github.com/adambarber/gorefuse"
	"gopkg.in/yaml.v3"
)

func yamlSetExtra(node *yaml.Node, extras reflect.Value, key keySegment, file FilePath, path FieldPath) error {
	if extras.Kind() != reflect.Map {
		return refuseAt(ErrUnknownField, file, gorefuse.Line(node.Line),
			gorefuse.Subject(`"`+string(key)+`"`),
			gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
	}

	if extras.IsNil() && extras.CanSet() {
		extras.Set(reflect.MakeMap(extras.Type()))
	}

	elem := reflect.New(extras.Type().Elem()).Elem()
	if err := yamlPopulateNode(node, elem, file, path); err != nil {
		return err
	}

	extras.SetMapIndex(reflect.ValueOf(string(key)).Convert(extras.Type().Key()), elem)

	return nil
}

func yamlNodeDecodeRefusal(file FilePath, node *yaml.Node, dst reflect.Value, path FieldPath, _ error) error {
	return typeMismatchRefusal(file, node.Line, ObjectKey(path), kindOf(dst.Type()), shapeUserTerm(node))
}

func yamlDuplicateKeyRefusal(file FilePath, key ObjectKey, first int) error {
	return refuseAt(ErrDuplicateKey, file, gorefuse.Line(first),
		gorefuse.Subject(`"`+string(key)+`"`),
		gorefuse.Why(`delete one of the two "`+string(key)+`" lines.`))
}
