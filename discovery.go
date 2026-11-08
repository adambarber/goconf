package goconf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adambarber/gorefuse"
)

// configName is the caller-declared name of a config: files are searched
// as NAME.EXT in each directory.
type configName string

// FoundFile is one config file discovery located, with its format and
// the search leg (directory) it was located in.
type FoundFile struct {
	Path   FilePath
	Format Format
	Leg    DirPath
}

// Discovery is the search the loader performs: the directories to look in
// (in order, later wins), the config name, and the required format.
type Discovery struct {
	Name   configName
	Format Format
	Search []DirPath
}

// discover lists the config files found in the search paths, in order:
// each directory holds at most one format's candidate unless the format
// narrows the search; zero candidates across all paths is refused.
func discover(d Discovery) ([]FoundFile, error) {
	if err := d.Format.check(); err != nil {
		return nil, err
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
		return nil, refuseWhy(ErrNothingFound,
			gorefuse.Subject("a file named \""+string(d.Name)+"\" with an extension some format spells"),
			gorefuse.Why("no candidate was found in any of the searched paths: "+string(renderPaths(d.Search))+"; create the file or widen the search."))
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
		return nil, refuseCause(
			gorefuse.At(gorefuse.File(dir), 0,
				gorefuse.Subject("the searched directory"),
				gorefuse.Why("discovery could not list this directory.")), err)
	}

	cands := candidateExts(entries, name)
	if required.IsUndetermined() && len(cands) > 1 {
		return nil, refuseWhy(ErrAmbiguous,
			gorefuse.Subject("\""+string(renderFiles(name, cands))+"\" in "+string(dir)),
			gorefuse.Why("one directory holds candidates with different-format extensions; declare the format so discovery never guesses."))
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

// checkCandidateName refuses a config name that could escape the search.
func checkCandidateName(name configName) error {
	n := string(name)
	if n == "" || filepath.IsAbs(n) || strings.ContainsAny(n, `\`) || strings.Contains(n, "..") || strings.Contains(n, "/") {
		return refuseWhy(ErrEscapingName,
			gorefuse.Subject(fmt.Sprintf("config name %q", n)),
			gorefuse.Why("the name must be a bare file name with no path separators or '..'; it would place the candidate file outside the searched directories."))
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
		return []FoundFile{{Path: p, Format: required, Leg: dir}}, nil
	}

	if errors.Is(statErr, fs.ErrNotExist) {
		return nil, nil
	}

	// any other stat failure is a real fault in the searched directory,
	// not absence: a load that could not look did not look
	return nil, refuseCause(
		gorefuse.At(gorefuse.File(p), 0,
			gorefuse.Subject("the candidate file"),
			gorefuse.Why("the searched directory could not be stat-ed; the fault is real, not absence.")), statErr)
}

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
			return nil, refuseWhy(ErrUnknownCandidate,
				gorefuse.Subject("\""+string(p)+"\""),
				gorefuse.Why("the candidate file's extension spells no format; rename the file or remove it."))
		}

		out = append(out, FoundFile{Path: p, Format: format, Leg: dir})
	}

	return out, nil
}
