package goconf

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

type sourceFile string

type collectorName string

type flowFinding string

// presenceTargetViolations reads one file and reports collectors that fail to
// hand their own reflect.Type parameter to canonicalizeMetaPaths. BPL cannot
// express argument flow: call() sees no arguments, type() sees no call site.
func presenceTargetViolations(t *testing.T, path sourceFile, collector collectorName) []flowFinding {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), string(path), nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	out := []flowFinding{}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == string(collector) {
			if finding := collectorFinding(fn); finding != "" {
				out = append(out, finding)
			}
		}
	}

	return out
}

// collectorFinding reports one collector's target-flow defect, or nothing.
func collectorFinding(fn *ast.FuncDecl) flowFinding {
	target := reflectTypeParam(fn)
	if target == nil {
		return flowFinding(fn.Name.Name + ": collector has no reflect.Type parameter")
	}

	called, forwarded := presenceHelperCall(fn.Body, target)

	switch {
	case !called:
		return flowFinding(fn.Name.Name + ": collector never calls canonicalizeMetaPaths")
	case !forwarded:
		return flowFinding(fn.Name.Name + ": canonicalizeMetaPaths does not receive " + target.Name)
	}

	return ""
}

// reflectTypeParam returns the first reflect.Type parameter's identifier.
func reflectTypeParam(fn *ast.FuncDecl) *ast.Ident {
	if fn.Type.Params == nil {
		return nil
	}

	for _, field := range fn.Type.Params.List {
		sel, ok := field.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Type" {
			continue
		}

		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "reflect" || len(field.Names) == 0 {
			continue
		}

		return field.Names[0]
	}

	return nil
}

// presenceHelperCall reports whether the body calls the shared helper and
// whether any such call forwards the named target parameter.
func presenceHelperCall(body *ast.BlockStmt, target *ast.Ident) (bool, bool) {
	called := false
	forwarded := false

	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		fn, ok := call.Fun.(*ast.Ident)
		if !ok || fn.Name != "canonicalizeMetaPaths" {
			return true
		}

		called = true

		if len(call.Args) != 2 {
			return true
		}

		arg, ok := call.Args[1].(*ast.Ident)

		forwarded = forwarded || (ok && arg.Name == target.Name)

		return true
	})

	return called, forwarded
}

// TestPresenceCollectorsReceiveTarget pins every production collector to the
// shared helper and to its own decode target parameter.
func TestPresenceCollectorsReceiveTarget(t *testing.T) {
	cases := map[sourceFile]collectorName{
		"meta_collect_json.go": "collectJSONMeta",
		"meta_collect_yaml.go": "collectYAMLMeta",
		"meta_collect_toml.go": "collectTOMLMeta",
	}

	for path, collector := range cases {
		t.Run(string(collector), func(t *testing.T) {
			if got := presenceTargetViolations(t, path, collector); len(got) != 0 {
				t.Errorf("presence target flow: %v", got)
			}
		})
	}
}

// TestPresenceTargetFlowFixturesProveItBites pins the checker: the positive
// fixture must be refused and the negative one must pass.
func TestPresenceTargetFlowFixturesProveItBites(t *testing.T) {
	positive := presenceTargetViolations(t,
		"policy/fixtures/presence_target_positive/fixture.go", "collectFixtureMeta")
	if len(positive) == 0 {
		t.Fatal("the positive fixture must be refused; the checker enforces nothing")
	}

	negative := presenceTargetViolations(t,
		"policy/fixtures/presence_target_negative/fixture.go", "collectFixtureMeta")
	if len(negative) != 0 {
		t.Fatalf("the negative fixture must pass: %v", negative)
	}
}
