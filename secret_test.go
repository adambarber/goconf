package goconf

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

type (
	secretText string
	secretKey  string
	safeText   string
)

type secretCfg struct {
	Credentials map[secretKey]secretText `json:"credentials" secret:"true" toml:"credentials" yaml:"credentials"`
	Region      safeText                 `json:"region"      toml:"region" yaml:"region"`
	Password    secretText               `json:"password"    secret:"true" toml:"password"    yaml:"password"`
	APIKey      secretText               `json:"api_key"     secret:"true" toml:"api_key"     yaml:"api_key"`
}

type unmarkedCfg struct {
	Region   safeText   `json:"region"   toml:"region"   yaml:"region"`
	Password secretText `json:"password" toml:"password" yaml:"password"`
}

type clearedCfg struct {
	Token secretText `json:"token" secret:"false" toml:"token" yaml:"token"`
}

type plainCfg struct {
	Region safeText `json:"region" toml:"region" yaml:"region"`
	Port   int      `json:"port"   toml:"port"   yaml:"port"`
}

func TestSecretGateRefusesUnmarkedKey(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameTOML: SourceBytes("region = \"eu\"\npassword = \"hunter2\"\n"),
	})

	_, _, err := Load[unmarkedCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if !errors.Is(err, ErrNeedDeclaration) {
		t.Fatalf("err = %v, want ErrNeedDeclaration", err)
	}
}

func TestSecretGateGoldenRefusal(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameTOML: SourceBytes("password = \"hunter2\"\n"),
	})

	_, _, err := Load[unmarkedCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err == nil {
		t.Fatal("unmarked secret key must be refused")
	}

	path := filepath.Join(string(root), string(dirLocal), string(fileNameTOML))

	want := "Error: " + path + `:1: undeclared secret: "password". add a secret tag to the field that holds it.`
	if err.Error() != want {
		t.Fatalf("render = %q, want %q", err.Error(), want)
	}
}

func TestSecretGateAcceptsDeclaredKeys(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameTOML: SourceBytes("password = \"p\"\napi_key = \"k\"\nregion = \"eu\"\n"),
	})

	got, _, err := Load[secretCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Password != "p" || got.APIKey != "k" {
		t.Fatalf("declared fields lost: %+v", got)
	}
}

func TestSecretGateAcceptsClearedKey(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameTOML: SourceBytes("token = \"t\"\n"),
	})

	got, _, err := Load[clearedCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil || got.Token != "t" {
		t.Fatalf("cleared key refused or lost: %+v, %v", got, err)
	}
}

func TestSecretGateInheritsParentMark(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameTOML: SourceBytes("[credentials]\npassword = \"x\"\n"),
	})

	_, _, err := Load[secretCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil {
		t.Fatalf("marked parent must cover dynamic children: %v", err)
	}
}

func TestFormatRedactsMarkedFields(t *testing.T) {
	cfg := secretCfg{
		Region:      "eu",
		Password:    "hunter2",
		APIKey:      "k",
		Credentials: map[secretKey]secretText{"password": "x"},
	}

	got := string(Redact(cfg))
	if strings.Contains(got, "hunter2") {
		t.Fatalf("secret leaked into render: %s", got)
	}

	if !strings.Contains(got, "Region:eu") {
		t.Fatalf("unmarked field missing from render: %s", got)
	}

	if strings.Count(got, "[redacted]") != 3 {
		t.Fatalf("redaction count = %d, want 3: %s", strings.Count(got, "[redacted]"), got)
	}
}

func TestFormatMatchesFmtForPointerFields(t *testing.T) {
	cfg := struct {
		Child *plainCfg
		Kids  []*plainCfg
	}{Child: &plainCfg{Region: "eu", Port: 8080}, Kids: []*plainCfg{{Region: "us", Port: 9090}}}

	if got, want := string(Redact(cfg)), fmt.Sprintf("%+v", cfg); got != want {
		t.Fatalf("Redact = %q, want %q", got, want)
	}
}

func TestFormatMatchesFmtForUnmarkedConfig(t *testing.T) {
	cfg := plainCfg{Region: "eu", Port: 8080}

	if got, want := string(Redact(cfg)), fmt.Sprintf("%+v", cfg); got != want {
		t.Fatalf("Redact = %q, want %q", got, want)
	}
}

func FuzzFormatRedaction(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("hunter2"))
	f.Add([]byte("[redacted]"))
	f.Add([]byte("a\nb"))

	f.Fuzz(func(t *testing.T, value []byte) {
		planted := secretText("LEAK-" + string(value))

		cfg := secretCfg{
			Region:      "eu",
			Password:    planted,
			APIKey:      planted,
			Credentials: map[secretKey]secretText{},
		}
		out := string(Redact(cfg))

		if strings.Contains(out, string(planted)) {
			t.Fatalf("Redact echoed a marked value: %s", out)
		}
	})
}
