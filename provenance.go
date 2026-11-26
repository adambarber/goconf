package goconf

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/adambarber/gorefuse"
)

// Provenance records one source that contributed to a load: where its
// bytes came from and the digest of those bytes.
type Provenance struct {
	Digest SourceDigest
	Path   FilePath
}

// SourceDigest is the hex sha256 of one source's bytes.
type SourceDigest string

// digestOf computes the digest of one source's bytes.
func digestOf(data SourceBytes) SourceDigest {
	sum := sha256.Sum256(data)
	return SourceDigest(hex.EncodeToString(sum[:]))
}

// provenanceOf builds the provenance entry for one located file.
func provenanceOf(f FoundFile, data SourceBytes) Provenance {
	return Provenance{Digest: digestOf(data), Path: f.Path}
}

// readSources reads every located file, in discovery order, refusing a read
// failure with the full path. os.Root enforces symlink confinement on the
// descriptor it reads, so a swap between check and read cannot escape.
func readSources(files []FoundFile, limit byteCap) ([]SourceBytes, []Provenance, error) {
	sources := make([]SourceBytes, 0, len(files))
	prov := make([]Provenance, 0, len(files))

	for _, f := range files {
		root, err := os.OpenRoot(string(f.Leg))
		if err != nil {
			return nil, nil, ioReadRefusal(f.Path, err)
		}

		raw, err := readConfined(root, f, limit)
		closeErr := root.Close()
		_ = closeErr

		if err != nil {
			return nil, nil, err
		}

		sources = append(sources, raw)
		prov = append(prov, provenanceOf(f, raw))
	}

	return sources, prov, nil
}

// readConfined measures and reads one candidate through a root-bound handle.
func readConfined(root *os.Root, f FoundFile, limit byteCap) (SourceBytes, error) {
	if err := confine(f.Path); err != nil {
		return nil, err
	}

	name := FileName(filepath.Base(string(f.Path)))

	info, err := root.Stat(string(name))
	if err != nil {
		return nil, confinedReadError(f, err)
	}

	if capErr := checkByteCap(int(info.Size()), f.Path, limit); capErr != nil {
		return nil, capErr
	}

	raw, err := root.ReadFile(string(name))
	if err != nil {
		return nil, confinedReadError(f, err)
	}

	return raw, nil
}

// confinedReadError keeps the user-facing confinement refusal when the root
// rejects a path; a genuine IO fault keeps the plain read-fault text.
func confinedReadError(f FoundFile, err error) error {
	if confErr := confine(f.Path); confErr != nil {
		return confErr
	}

	return ioReadRefusal(f.Path, err)
}

// confine names the refusal for a candidate whose symlink target is absolute
// or resolves outside the searched directory. os.Root enforces the boundary;
// this produces the reason and the target the user needs to see.
func confine(p FilePath) error {
	file := string(p)

	if link, err := os.Readlink(file); err == nil && filepath.IsAbs(link) {
		return refuseWhy(ErrOutsideSearch,
			gorefuse.Subject("\""+file+"\" which points at \""+link+"\""),
			gorefuse.Why("the symlink target must be relative and stay inside the searched directory."))
	}

	resolved, err := filepath.EvalSymlinks(file)
	if err != nil {
		return refuseIO(
			gorefuse.At(gorefuse.File(p), 0,
				gorefuse.Subject("could not read the file"),
				gorefuse.Why("check that the file and its link target exist and are readable.")), err)
	}

	dir := filepath.Dir(file)

	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return refuseIO(
			gorefuse.At(gorefuse.File(DirPath(dir)), 0,
				gorefuse.Subject("could not read the directory"),
				gorefuse.Why("check the directory's permissions (it must be readable by your user).")), err)
	}

	if !withinDir(resolvedPath(resolved), resolvedPath(resolvedDir)) {
		return refuseWhy(ErrOutsideSearch,
			gorefuse.Subject("\""+file+"\" which points at \""+resolved+"\""),
			gorefuse.Why("the symlink-resolved target leaves the searched directory; keep the target inside the searched tree."))
	}

	return nil
}

// resolvedPath is a symlink-resolved absolute file path.
type resolvedPath string

// ioReadRefusal builds the read-fault refusal: the path appears once (in
// the frame), the cause tail carries the OS detail, and the fix is one
// user action — never a search leg or a fault-vs-absence lecture.
func ioReadRefusal(p FilePath, err error) error {
	return refuseIO(
		gorefuse.At(gorefuse.File(p), 0,
			gorefuse.Subject("could not read the file"),
			gorefuse.Why("check the file's permissions (it must be readable by your user).")), err)
}

// withinDir compares whole path segments, so a sibling named "..foo" stays
// inside while ".." and "../" escape.
func withinDir(resolved, searched resolvedPath) bool {
	rel, err := filepath.Rel(string(searched), string(resolved))
	if err != nil {
		return false
	}

	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// defaultSearchPaths is the discovery default: the user's global config
// directory first, then the local file of the same name in the working
// directory; local wins over global.
func defaultSearchPaths(s *settings, name configName) []DirPath {
	global := s.globalDir
	if global == "" {
		if dir, err := os.UserConfigDir(); err == nil {
			global = DirPath(filepath.Join(dir, string(name)))
		}
	}

	if global == "" {
		return []DirPath{DirPath(".")}
	}

	return []DirPath{global, DirPath(".")}
}
