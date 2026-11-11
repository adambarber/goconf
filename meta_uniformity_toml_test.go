package goconf

import "testing"

func TestMetaArrayUniformityTOMLAOT(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("[[items]]\nname = \"a\"\n\n[[items]]\nname = \"b\"\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Items []muItem `toml:"items"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertArrayOfObjectsShape(t, meta)
}

func TestMetaArrayUniformityTOMLInlineScalar(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("nums = [1, 2, 3]\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Nums []int `toml:"nums"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertArrayContainerAndElements(t, meta, "nums", 3, false)
}

func TestMetaArrayUniformityTOMLInlineTableArray(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("items = [{name=\"a\"}, {name=\"b\"}]\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Items []muItem `toml:"items"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertArrayOfObjectsShape(t, meta)
}

func assertArrayContainerAndElements(t *testing.T, meta Meta, container OccurrencePath, count int, requireContainer bool) {
	t.Helper()

	if requireContainer {
		cs := meta.ByPath(container)
		if len(cs) == 0 {
			t.Fatalf("missing container %q: %#v", container, meta.Entries())
		}

		if cs[0].Key != OccurrenceKey(container) {
			t.Fatalf("container key=%q want=%q entry=%#v", cs[0].Key, container, cs[0])
		}
	}

	for i := range count {
		p := OccurrencePath(string(container) + "[" + string(itoa(i)) + "]")
		es := meta.ByPath(p)

		if len(es) == 0 {
			t.Fatalf("missing element %q: %#v", p, meta.Entries())
		}

		wantKey := OccurrenceKey("[" + string(itoa(i)) + "]")
		if es[0].Key != wantKey {
			t.Fatalf("element %q key=%q want=%q entry=%#v", p, es[0].Key, wantKey, es[0])
		}
	}
}

func assertArrayOfObjectsShape(t *testing.T, meta Meta) {
	t.Helper()

	for i := range 2 {
		assertArrayObjectElementShape(t, meta, "items", OccurrenceKey("name"), i)
	}
}

func assertArrayObjectElementShape(t *testing.T, meta Meta, container OccurrencePath, child OccurrenceKey, i int) {
	t.Helper()

	elemPath := OccurrencePath(string(container) + "[" + string(itoa(i)) + "]")
	es := meta.ByPath(elemPath)

	if len(es) == 0 {
		t.Fatalf("missing element %q: %#v", elemPath, meta.Entries())
	}

	wantKey := OccurrenceKey("[" + string(itoa(i)) + "]")
	if es[0].Key != wantKey {
		t.Fatalf("element %q key=%q want=%q entry=%#v", elemPath, es[0].Key, wantKey, es[0])
	}

	childPath := OccurrencePath(string(elemPath) + "." + string(child))
	cs := meta.ByPath(childPath)

	if len(cs) == 0 {
		t.Fatalf("missing child %q: %#v", childPath, meta.Entries())
	}

	if cs[0].Key != child {
		t.Fatalf("child %q key=%q want=%q", childPath, cs[0].Key, child)
	}

	elemIdx := metaIndexOfPath(meta, elemPath)
	childIdx := metaIndexOfPath(meta, childPath)

	if elemIdx < 0 || childIdx < 0 || elemIdx > childIdx {
		t.Fatalf("order elem=%q(%d) child=%q(%d): %#v", elemPath, elemIdx, childPath, childIdx, meta.Entries())
	}
}

func metaIndexOfPath(meta Meta, path OccurrencePath) int {
	for i, e := range meta.Entries() {
		if e.Path == path {
			return i
		}
	}

	return -1
}

func itoa(i int) OccurrenceKey {
	if i == 0 {
		return "0"
	}

	if i == 1 {
		return "1"
	}

	if i == 2 {
		return "2"
	}

	return fmtIntKey(i)
}

func fmtIntKey(i int) OccurrenceKey {
	if i == 0 {
		return "0"
	}

	neg := false
	if i < 0 {
		neg = true
		i = -i
	}

	buf := [20]byte{}
	n := len(buf)

	for i > 0 {
		n--
		buf[n] = byte('0' + i%10)
		i /= 10
	}

	if neg {
		n--
		buf[n] = '-'
	}

	return OccurrenceKey(buf[n:])
}
