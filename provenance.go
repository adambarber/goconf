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

// readSources reads every located file, in discovery order, naming the
// full path and search leg on a read failure. Each file is confined at
// read time so a symlink cannot smuggle bytes in from outside.
func readSources(files []FoundFile) ([]SourceBytes, []Provenance, error) {
	sources := make([]SourceBytes, 0, len(files))

	prov := make([]Provenance, 0, len(files))
	for _, f := range files {
		if err := confine(f.Path); err != nil {
			return nil, nil, err
		}

		raw, err := os.ReadFile(string(f.Path))
		if err != nil {
			return nil, nil, refuseIO(
				gorefuse.At(gorefuse.File(f.Path), 0,
					gorefuse.Subject("the candidate file in search leg "+string(f.Leg)),
					gorefuse.Why("the file could not be read; the fault is real, not absence.")), err)
		}

		sources = append(sources, raw)
		prov = append(prov, provenanceOf(f, raw))
	}

	return sources, prov, nil
}

// confine refuses a candidate file whose symlink-resolved target leaves
// the directory that was searched, naming the file and the target.
func confine(p FilePath) error {
	file := string(p)

	resolved, err := filepath.EvalSymlinks(file)
	if err != nil {
		return refuseIO(
			gorefuse.At(gorefuse.File(p), 0,
				gorefuse.Subject("the candidate file"),
				gorefuse.Why("the symlink target could not be resolved.")), err)
	}

	dir := filepath.Dir(file)

	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return refuseIO(
			gorefuse.At(gorefuse.File(DirPath(dir)), 0,
				gorefuse.Subject("the searched directory"),
				gorefuse.Why("the directory could not be resolved.")), err)
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

// withinDir reports whether the resolved path stays inside the searched
// directory: the same directory, or one nested beneath it.
func withinDir(resolved, searched resolvedPath) bool {
	rel, err := filepath.Rel(string(searched), string(resolved))
	if err != nil {
		return false
	}

	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "..")
}

// defaultSearchPaths is the discovery default: the user's global config
// directory first, then the local file of the same name in the working
// directory; local wins over global.
func defaultSearchPaths(s settings, name configName) []DirPath {
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
