package fixtures

// One decision, twice — the directive this rule refuses.
func work() { //nolint:errcheck // a lint rule switched off in a comment
}
