package goconf

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

type inflectionCfg struct {
	Passwords    secretText `json:"passwords"     toml:"passwords"     yaml:"passwords"`
	Secrets      secretText `json:"secrets"       toml:"secrets"       yaml:"secrets"`
	Tokens       secretText `json:"tokens"        toml:"tokens"        yaml:"tokens"`
	CredentialsX secretText `json:"credentials_x" toml:"credentials_x" yaml:"credentials_x"`
	PassWord     secretText `json:"passWord"      toml:"passWord"      yaml:"passWord"`
	SecretKey    secretText `json:"secretKey"     toml:"secretKey"     yaml:"secretKey"`
}

type unexportedToken string

type unexportedSecret struct {
	token unexportedToken `secret:"true"`
}

func TestSecretGateRefusesInflectedNames(t *testing.T) {
	keys := []string{"passwords", "secrets", "tokens", "credentials_x", "passWord", "secretKey"}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			_, _, err := LoadBytes[inflectionCfg](SourceBytes("{"+strconv.Quote(key)+":\"v\"}"),
				Spec{Name: configName("app"), Format: FormatJSON})
			if !errors.Is(err, ErrNeedDeclaration) {
				t.Fatalf("err = %v, want ErrNeedDeclaration", err)
			}
		})
	}
}

func TestSecretGateAcrossFormats(t *testing.T) {
	cases := []struct {
		format Format
		body   SourceBytes
	}{
		{FormatJSON, SourceBytes(`{"password":"v"}`)},
		{FormatYAML, SourceBytes("password: v\n")},
		{FormatTOML, SourceBytes("password = \"v\"\n")},
	}

	for _, tc := range cases {
		t.Run(string(tc.format), func(t *testing.T) {
			_, _, err := LoadBytes[unmarkedCfg](tc.body, Spec{Name: configName("app"), Format: tc.format})
			if !errors.Is(err, ErrNeedDeclaration) {
				t.Fatalf("err = %v, want ErrNeedDeclaration", err)
			}
		})
	}
}

func TestSecretGateAcrossSources(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("region = \"eu\"\n")},
		dirLocal:  {fileNameTOML: SourceBytes("password = \"v\"\n")},
	})

	_, _, err := Load[unmarkedCfg](envSpec(), WithSearchPaths(legPaths(root, dirGlobal, dirLocal)...))
	if !errors.Is(err, ErrNeedDeclaration) {
		t.Fatalf("err = %v, want ErrNeedDeclaration", err)
	}
}

func TestRedactMarkedUnexportedField(t *testing.T) {
	got := string(Redact(unexportedSecret{token: "hunter2"}))

	if strings.Contains(got, "hunter2") || !strings.Contains(got, "[redacted]") {
		t.Fatalf("render = %q, want marked unexported field redacted", got)
	}
}
