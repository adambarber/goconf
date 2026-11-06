package goconf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
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

// readSources reads every located file, in discovery order.
func readSources(files []FoundFile) ([]SourceBytes, []Provenance, error) {
	sources := make([]SourceBytes, 0, len(files))

	prov := make([]Provenance, 0, len(files))
	for _, f := range files {
		raw, err := os.ReadFile(string(f.Path))
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", f.Path, err)
		}

		sources = append(sources, raw)
		prov = append(prov, provenanceOf(f, raw))
	}

	return sources, prov, nil
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
