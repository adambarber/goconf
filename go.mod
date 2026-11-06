module github.com/adambarber/goconf

go 1.27

// Degraded-network workaround: local copies fetched over HTTPS while the
// host's direct module fetch was failing; replace with a plain `go get` +
// `go mod tidy` when the connection recovers.
replace github.com/BurntSushi/toml => ./deps/src/github.com/BurntSushi/toml@v1.4.0

replace gopkg.in/yaml.v3 => ./deps/src/gopkg.in/yaml.v3@v3.0.1

require (
	github.com/BurntSushi/toml v1.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
