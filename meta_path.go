package goconf

import "strconv"

// metaSeg is a path/key segment text used in metadata path construction.
type metaSeg string

func appendDotSeg(parent, seg metaSeg) metaSeg {
	if parent == "" {
		return seg
	}

	return metaSeg(string(parent) + "." + string(seg))
}

func occPath(parent OccurrencePath, seg metaSeg) OccurrencePath {
	return OccurrencePath(appendDotSeg(metaSeg(parent), seg))
}

func occIndex(parent OccurrencePath, index int) OccurrencePath {
	return OccurrencePath(string(parent) + "[" + strconv.Itoa(index) + "]")
}

func schemaPath(parent SchemaPath, seg metaSeg) SchemaPath {
	return SchemaPath(appendDotSeg(metaSeg(parent), seg))
}

func schemaIndex(parent SchemaPath) SchemaPath { return parent }
