package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// handlerCases reads the string literals of every case in one function. The
// point is to compare the panel router with the token router mechanically:
// a panel action added without a v2 case fails this test instead of being
// discovered by a caller whose token suddenly gets "unknown action".
func handlerCases(t *testing.T, file, function string) map[string]bool {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	cases := map[string]bool{}
	found := false
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != function {
			continue
		}
		found = true
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			clause, ok := node.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, expr := range clause.List {
				if literal, ok := expr.(*ast.BasicLit); ok && literal.Kind == token.STRING {
					value, err := strconv.Unquote(literal.Value)
					if err == nil {
						cases[value] = true
					}
				}
			}
			return true
		})
	}
	if !found {
		t.Fatalf("%s has no function %s", file, function)
	}
	return cases
}

func TestAPIv2CoversEveryPanelAction(t *testing.T) {
	panelPost := handlerCases(t, "apiHandler.go", "postHandler")
	panelGet := handlerCases(t, "apiHandler.go", "getHandler")
	tokenPost := handlerCases(t, "apiV2Handler.go", "postHandler")
	tokenGet := handlerCases(t, "apiV2Handler.go", "getHandler")

	// Session-only actions: a token has no browser session to start or end.
	sessionOnlyPost := map[string]bool{"login": true}
	sessionOnlyGet := map[string]bool{"logout": true}

	for action := range panelPost {
		if sessionOnlyPost[action] {
			if tokenPost[action] {
				t.Errorf("session-only POST %q must not be reachable with a token", action)
			}
			continue
		}
		if !tokenPost[action] {
			t.Errorf("panel POST action %q is missing from the token router", action)
		}
	}
	for action := range panelGet {
		if sessionOnlyGet[action] {
			if tokenGet[action] {
				t.Errorf("session-only GET %q must not be reachable with a token", action)
			}
			continue
		}
		if !tokenGet[action] {
			t.Errorf("panel GET action %q is missing from the token router", action)
		}
	}
}
