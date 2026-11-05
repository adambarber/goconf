package "goconf"

rule branching_budget {
    fix      = "break the decision tree apart"
    lang     = go
    max      = 10
    negative = { file = "policy/fixtures/branching_negative/fixture.go", expects = no_finding }
    owner    = "platform"
    positive = { file = "policy/fixtures/branching_positive/fixture.go", expects = finding(rule: "branching_budget", marker: "branches 11 ways, over the 10 allowed"), count = 1 }
    type     = budget
    unit     = branching
    why      = "a function that can go more than ten ways is a function nobody holds in their head at once; each fork, case, and short-circuit is a path a reader must keep open"
}

rule comment_budget {
    fix      = "compress it, or make it a document in docs/ where it gets an index"
    lang     = go
    max      = 3
    negative = { file = "policy/fixtures/comment_budget_negative/fixture.go", expects = no_finding }
    owner    = "platform"
    positive = { file = "policy/fixtures/comment_budget_positive/fixture.go", expects = finding(rule: "comment_budget", marker: "4 lines of prose, over the 3 allowed"), count = 1 }
    type     = budget
    unit     = comment_lines
    why      = "a comment longer than three lines is documentation wearing a comment's clothes, and the place for it is docs/, where it gets an index"
}

rule file_budget {
    fix      = "split it by concern"
    lang     = go
    max      = 200
    negative = { file = "policy/fixtures/file_budget_negative/fixture.go", expects = no_finding }
    owner    = "platform"
    positive = { file = "policy/fixtures/file_budget_positive/fixture.go", expects = finding(rule: "file_budget", marker: "is 202 lines, over the 200 allowed"), count = 1 }
    type     = budget
    unit     = lines
    why      = "a file past two hundred code lines has more than one concern in it, and the split never happens on its own"
}

rule function_budget {
    fix      = "split it by concern"
    lang     = go
    max      = 60
    negative = { file = "policy/fixtures/function_lines_negative/fixture.go", expects = no_finding }
    owner    = "platform"
    positive = { file = "policy/fixtures/function_lines_positive/fixture.go", expects = finding(rule: "function_budget", marker: "func longFn is 61 lines, over the 60 allowed"), count = 1 }
    type     = budget
    unit     = function_lines
    why      = "a function over sixty code lines is a procedure hiding its seams; the split never happens on its own, and the count is the function's own, not the file's"
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
