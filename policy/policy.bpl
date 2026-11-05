package "goconf"

rule no_clones {
    fix      = "extract the shared decision, or spell the difference the copy hides"
    lang     = any
    match    = clone(lines: 3, tokens: 20)
    negative = { file = "policy/fixtures/no_clones_negative/fixture.go", expects = no_finding }
    owner    = "platform"
    positive = { file = "policy/fixtures/no_clones_positive/fixture.go", expects = finding(rule: "no_clones", marker: "two copies of one decision"), count = 1 }
    severity = error
    why      = "a clone is one decision written twice, and they drift the first time one of them changes"
}
