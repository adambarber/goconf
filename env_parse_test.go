package goconf

import (
	"errors"
	"strings"
	"testing"
)

type envKey string

type envWantPhrase string

type envNumericCfg struct {
	Count uint16  `json:"count" toml:"count" yaml:"count"`
	Ratio float64 `json:"ratio" toml:"ratio" yaml:"ratio"`
}

type envListCfg struct {
	Hosts []envText `json:"hosts" toml:"hosts" yaml:"hosts"`
}

type envLabelsCfg struct {
	Labels map[envKey]envText `json:"labels" toml:"labels" yaml:"labels"`
}

type envUnsupportedCfg struct {
	Cplx complex128 `json:"cplx" toml:"cplx" yaml:"cplx"`
}

const (
	envProbeCount = "42"
	envProbeRatio = "1.5"
	envBadNumber  = "not-a-number"
)

// TestEnvParsesUnsignedAndFloat drives the numeric env setters through a real
// load and pins the parsed values.
func TestEnvParsesUnsignedAndFloat(t *testing.T) {
	t.Setenv("APP_COUNT", envProbeCount)
	t.Setenv("APP_RATIO", envProbeRatio)

	got, _, err := LoadBytes[envNumericCfg]([]byte(""), envSpec(), WithEnv("APP"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Count != 42 || got.Ratio != 1.5 {
		t.Fatalf("env numeric values = %d/%v, want 42/1.5", got.Count, got.Ratio)
	}
}

// TestEnvRefusesInvalidNumericValues pins that bad numeric env values are
// refusals naming the expected shape, not silent zeros.
func TestEnvRefusesInvalidNumericValues(t *testing.T) {
	cases := map[string]string{
		"APP_COUNT": envBadNumber,
		"APP_RATIO": envBadNumber,
	}

	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, value)

			_, _, err := LoadBytes[envNumericCfg]([]byte(""), envSpec(), WithEnv("APP"))
			if !errors.Is(err, ErrEnvParse) {
				t.Fatalf("err = %v, want ErrEnvParse", err)
			}
		})
	}
}

func assertEnvRefusal(t *testing.T, load func() error, want envWantPhrase) {
	t.Helper()

	err := load()
	if !errors.Is(err, ErrEnvParse) {
		t.Fatalf("err = %v, want ErrEnvParse", err)
	}

	if !strings.Contains(err.Error(), string(want)) {
		t.Fatalf("err = %q, want shape %q", err, want)
	}
}

// TestEnvRefusesInvalidCompositeValues pins the expected-shape wording for
// list and map env fields.
func TestEnvRefusesInvalidCompositeValues(t *testing.T) {
	t.Run("APP_HOSTS", func(t *testing.T) {
		t.Setenv("APP_HOSTS", "not-json")

		assertEnvRefusal(t, func() error {
			_, _, err := LoadBytes[envListCfg]([]byte(""), envSpec(), WithEnv("APP"))
			return err
		}, "JSON for a list")
	})

	t.Run("APP_LABELS", func(t *testing.T) {
		t.Setenv("APP_LABELS", "not-json")

		assertEnvRefusal(t, func() error {
			_, _, err := LoadBytes[envLabelsCfg]([]byte(""), envSpec(), WithEnv("APP"))
			return err
		}, "JSON for a map")
	})
}

// TestEnvRefusesUnsupportedKind pins the refusal for a field shape env cannot
// parse at all.
func TestEnvRefusesUnsupportedKind(t *testing.T) {
	t.Setenv("APP_CPLX", "1")

	_, _, err := LoadBytes[envUnsupportedCfg]([]byte(""), envSpec(), WithEnv("APP"))
	if !errors.Is(err, ErrEnvParse) {
		t.Fatalf("err = %v, want ErrEnvParse", err)
	}

	if !strings.Contains(err.Error(), "a supported value") {
		t.Fatalf("err = %q, want supported-value wording", err)
	}
}

// FuzzEnvNumericParse drives arbitrary env values through uint and float
// parsing and asserts the refusal never echoes the value.
func FuzzEnvNumericParse(f *testing.F) {
	for _, seed := range []string{"", "42", "1.5", "-1", "abc", "999999999999999999999"} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, value []byte) {
		planted := "ENV-" + strings.ReplaceAll(string(value), "\x00", "x")
		t.Setenv("APP_COUNT", planted)
		t.Setenv("APP_RATIO", planted)

		_, _, err := LoadBytes[envNumericCfg]([]byte(""), envSpec(), WithEnv("APP"))
		if err == nil {
			return
		}

		if strings.Contains(err.Error(), planted) {
			t.Fatalf("env value leaked: %s", err.Error())
		}
	})
}
