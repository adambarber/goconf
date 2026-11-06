package goconf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// configName is the caller-declared name of a config: files are searched
// as NAME.EXT in each directory.
type configName string

// FoundFile is one config file discovery located, with its format.
type FoundFile struct {
	Path   FilePath
	Format Format
}

// Discovery is the search the loader performs: the directories to look in
// (in order, later wins), the config name, and the required format.
type Discovery struct {
	Name   configName
	Format Format
	Search []DirPath
}

// errAmbiguous is the refusal for one directory holding candidates with
// different-format extensions; discovery never guesses.
var errAmbiguous = errors.New("ambiguous config candidates with different formats")

// errNothingFound is the refusal for a discovery that found no candidate
// anywhere: a load that found nothing loaded nothing.
var errNothingFound = errors.New("no config file found in any searched path")

// discover lists the config files found in the search paths, in order:
// each directory holds at most one format's candidate unless the format
// narrows the search; zero candidates across all paths is refused.
func discover(d Discovery) ([]FoundFile, error) {
	if err := d.Format.check(); err != nil {
		return nil, fmt.Errorf("discover: %w", err)
	}

	if err := checkCandidateName(d.Name); err != nil {
		return nil, err
	}

	found := []FoundFile{}

	for _, dir := range d.Search {
		one, err := discoverDir(dir, d.Name, d.Format)
		if err != nil {
			return nil, err
		}

		found = append(found, one...)
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("discover: %w in %s", errNothingFound, renderPaths(d.Search))
	}

	return found, nil
}

// discoverDir finds candidates in one directory, refusing ambiguity.
func discoverDir(dir DirPath, name configName, required Format) ([]FoundFile, error) {
	entries, err := os.ReadDir(string(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("discover: %w", err)
	}

	cands := candidateExts(entries, name)
	if required.IsUndetermined() && len(cands) > 1 {
		return nil, fmt.Errorf("discover: %w in %s: %s", errAmbiguous, dir, renderFiles(name, cands))
	}

	if !required.IsUndetermined() {
		return narrow(required, dir, name)
	}

	return onePer(cands, dir, name)
}

// candidateExts lists the extensions of files named NAME.EXT in entries.
func candidateExts(entries []os.DirEntry, name configName) []Extension {
	exts := []Extension{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := extOf(FileName(e.Name()), name)
		if ext != "" {
			exts = append(exts, ext)
		}
	}

	return exts
}

// extOf returns the extension of a file named NAME.EXT, or empty when
// the file is not a candidate: the extension must include its dot.
func extOf(file FileName, name configName) Extension {
	base := string(file)
	if !strings.HasPrefix(base, string(name)) || base == string(name) {
		return ""
	}

	ext := Extension(strings.TrimPrefix(base, string(name)))
	if !strings.HasPrefix(string(ext), ".") {
		return ""
	}

	return ext
}

// candidatePath builds one candidate file path from a searched directory
// and the config name.
func candidatePath(dir DirPath, name configName, ext Extension) (FilePath, error) {
	if err := checkCandidateName(name); err != nil {
		return "", err
	}

	return FilePath(filepath.Join(string(dir), string(name)+string(ext))), nil
}

// errEscapingName is the refusal for a config name that would place the
// candidate file outside the searched directories.
var errEscapingName = errors.New("config name must be a bare file name with no path separators or '..'; got")

// checkCandidateName refuses a config name that could escape the search.
func checkCandidateName(name configName) error {
	n := string(name)
	if n == "" || filepath.IsAbs(n) || strings.ContainsAny(n, `\`) || strings.Contains(n, "..") || strings.Contains(n, "/") {
		return fmt.Errorf("discover: %w %q", errEscapingName, n)
	}

	return nil
}

// narrow returns the single candidate a required format selects, if the
// file exists; a stat failure other than not-exists is refused, naming
// the candidate path, never treated as absence.
func narrow(required Format, dir DirPath, name configName) ([]FoundFile, error) {
	p, err := candidatePath(dir, name, Extension(required))
	if err != nil {
		return nil, err
	}

	_, statErr := os.Stat(string(p))
	if statErr == nil {
		return []FoundFile{{Path: p, Format: required}}, nil
	}

	if errors.Is(statErr, fs.ErrNotExist) {
		return nil, nil
	}

	// any other stat failure is a real fault in the searched directory,
	// not absence: a load that could not look did not look
	return nil, fmt.Errorf("discover: %s: %w", p, statErr)
}

// errUnknownCandidate is the refusal for a candidate file whose
// extension no format spells: discovery never silently skips a file the
// name boundary selected.
var errUnknownCandidate = errors.New("candidate file has an extension no format spells; the file must be renamed or removed")

// onePer builds one FoundFile per candidate extension, refusing any
// extension no format spells by naming the file.
func onePer(cands []Extension, dir DirPath, name configName) ([]FoundFile, error) {
	out := make([]FoundFile, 0, len(cands))
	for _, ext := range cands {
		p, err := candidatePath(dir, name, ext)
		if err != nil {
			return nil, err
		}

		format, ok := formatOf(ext)
		if !ok {
			return nil, fmt.Errorf("discover: %w: %s", errUnknownCandidate, p)
		}

		out = append(out, FoundFile{Path: p, Format: format})
	}

	return out, nil
}
