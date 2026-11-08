package "goconf"

ignore positiveFixtures {
    paths  = ["policy/fixtures/**"]
    reason = "a positive fixture must violate a rule to prove the rule bites; the fixture accounting, not the whole-tree run, is where a fixture's finding is judged"
    owner  = "platform"
}

ignore errorRendering {
    paths  = ["refusal.go"]
    reason = "the stdlib error interface's Error() string method set is the stdlib's spelling; the named-kind rule cannot apply to the two methods that satisfy it"
    owner  = "platform"
}

rule branching_budget {
    lang    = go
    type    = budget
    unit    = branching
    max     = 10
    why     = "a function that can go more than ten ways is a function nobody holds in their head at once; each fork, case, and short-circuit is a path a reader must keep open"
    fix     = "break the decision tree apart"
    owner   = "platform"
    positive = { file = "policy/fixtures/branching_positive/fixture.go", expects = finding(rule: "branching_budget", marker: "branches 11 ways, over the 10 allowed"), count = 1 }
    negative = { file = "policy/fixtures/branching_negative/fixture.go", expects = no_finding }
}

rule comment_budget {
    lang    = go
    type    = budget
    unit    = comment_lines
    max     = 3
    why     = "a comment longer than three lines is documentation wearing a comment's clothes, and the place for it is docs/, where it gets an index"
    fix     = "compress it, or make it a document in docs/ where it gets an index"
    owner   = "platform"
    positive = { file = "policy/fixtures/comment_budget_positive/fixture.go", expects = finding(rule: "comment_budget", marker: "4 lines of prose, over the 3 allowed"), count = 1 }
    negative = { file = "policy/fixtures/comment_budget_negative/fixture.go", expects = no_finding }
}

rule file_budget {
    lang    = go
    type    = budget
    unit    = lines
    max     = 200
    why     = "a file past two hundred code lines has more than one concern in it, and the split never happens on its own"
    fix     = "split it by concern"
    owner   = "platform"
    positive = { file = "policy/fixtures/file_budget_positive/fixture.go", expects = finding(rule: "file_budget", marker: "is 202 lines, over the 200 allowed"), count = 1 }
    negative = { file = "policy/fixtures/file_budget_negative/fixture.go", expects = no_finding }
}

rule function_budget {
    lang    = go
    type    = budget
    unit    = function_lines
    max     = 60
    why     = "a function over sixty code lines is a procedure hiding its seams; the split never happens on its own, and the count is the function's own, not the file's"
    fix     = "split it by concern"
    owner   = "platform"
    positive = { file = "policy/fixtures/function_lines_positive/fixture.go", expects = finding(rule: "function_budget", marker: "func longFn is 61 lines, over the 60 allowed"), count = 1 }
    negative = { file = "policy/fixtures/function_lines_negative/fixture.go", expects = no_finding }
}

rule no_untyped_value {
    lang    = go
    match    = type(position: "var", of: "string")
            or type(position: "var", of: "any")
            or type(position: "parameter", of: "string")
            or type(position: "parameter", of: "any")
            or type(position: "result", of: "string")
            or type(position: "result", of: "any")
            or type(position: "field", of: "string")
            or type(position: "field", of: "any")
            or type(position: "type_spec", of: "any")
            or type(position: "assertion", of: "any")
            or type(position: "type_argument", of: "any")
    why     = "a value whose type is spelled string is a bag of bytes, and an empty interface is a type nobody chose: name the domain kind"
    fix     = "declare a named type and use it in the type position"
    owner   = "platform"
    positive = { file = "policy/fixtures/no_untyped_value_positive/fixture.go", expects = finding(rule: "no_untyped_value", marker: "typed"), count = 3 }
    negative = { file = "policy/fixtures/no_untyped_value_negative/fixture.go", expects = no_finding }
}

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

# Doors: every gate a passing commit must survive is declared here, so the
# policy is the single ledger of what the commit path runs. The engine's
# verify pass covers the rules above; the doors below are the rest.

command gofumpt {
    run    = "golangci-lint fmt --diff"
    why    = "unformatted files are a formatting defect the diff exposes; the exit code is the verdict"
    owner  = "platform"
}

command lint {
    run    = "golangci-lint run ./..."
    why    = "the strict merged lint gate; any finding refuses the commit"
    owner  = "platform"
}

command test {
    run    = "go test -count=1 ./..."
    why    = "-count=1 refuses the cache: the suite must behave, not remember"
    owner  = "platform"
}
