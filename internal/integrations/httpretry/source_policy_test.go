package httpretry

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestProductionHTTPCallsUseReviewedPaths prevents a future adapter from
// bypassing the shared retry policy with a direct HTTP call or transport
// round trip. A reviewed exception must be added with its reason and evidence.
func TestProductionHTTPCallsUseReviewedPaths(t *testing.T) {
	root := moduleRoot(t)
	var violations []string
	observed := make(map[string][]string)

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "artifacts", "build", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		httpAliases := netHTTPAliases(file)
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}

		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			functionName := function.Name.Name
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if isHTTPExecutionMethod(selector.Sel.Name) {
					name := expressionString(t, fset, selector)
					key := filepath.ToSlash(relative) + ":" + functionName
					observed[key] = append(observed[key], name)
					return true
				}

				if alias, ok := selector.X.(*ast.Ident); ok && httpAliases[alias.Name] && isHTTPConvenienceMethod(selector.Sel.Name) {
					position := fset.Position(call.Pos())
					violations = append(violations, filepath.ToSlash(relative)+":"+strconv.Itoa(position.Line)+" calls net/http."+selector.Sel.Name+" directly")
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan production Go sources: %v", err)
	}

	// Provider API reads go through their adapter-specific classification and the
	// shared GET-only policy. OAuth token exchange remains a single-use POST; its
	// transport wrapper restricts it to the configured Google token endpoint.
	// The rate limiter delegates to its wrapped transport.
	approved := map[string][]string{
		"internal/integrations/bling/api_client.go:doFinancialResourceAttempt": {"c.httpClient.Do"},
		"internal/integrations/bling/api_resources.go:getDataRecordAttempt":    {"c.httpClient.Do"},
		"internal/integrations/bling/oauth.go:exchange":                        {"c.httpClient.Do"},
		"internal/integrations/bling/rate_limit.go:Do":                         {"d.next.Do"},
		"internal/integrations/bling/retry.go:withBlingReadRetry":              {"httpretry.Do"},
		"internal/integrations/googledrive/client.go:doGETAttempt":             {"c.httpClient.Do"},
		"internal/integrations/googledrive/client.go:withGETRetry":             {"httpretry.Do"},
		"internal/integrations/googledrive/auth.go:RoundTrip":                  {"t.base.RoundTrip"},
		"internal/ipc/namedpipe_windows.go:close":                              {"p.once.Do"},
	}
	for key, calls := range observed {
		want, ok := approved[key]
		if !ok {
			violations = append(violations, key+" has unreviewed calls: "+strings.Join(calls, ", "))
			continue
		}
		if !sameStrings(calls, want) {
			violations = append(violations, key+" calls "+strings.Join(calls, ", ")+"; want "+strings.Join(want, ", "))
		}
	}
	for key := range approved {
		if _, ok := observed[key]; !ok {
			violations = append(violations, key+" is no longer present; review the HTTP inventory")
		}
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("HTTP calls need a retry-safety review before merge:\n- %s", strings.Join(violations, "\n- "))
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate source-policy test file")
	}
	for directory := filepath.Dir(sourceFile); ; directory = filepath.Dir(directory) {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("go.mod not found above source-policy test")
		}
	}
}

func netHTTPAliases(file *ast.File) map[string]bool {
	aliases := make(map[string]bool)
	for _, imported := range file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil || path != "net/http" {
			continue
		}
		alias := "http"
		if imported.Name != nil {
			alias = imported.Name.Name
		}
		aliases[alias] = true
	}
	return aliases
}

func isHTTPConvenienceMethod(name string) bool {
	switch name {
	case "Get", "Head", "Post", "PostForm":
		return true
	default:
		return false
	}
}

func isHTTPExecutionMethod(name string) bool {
	switch name {
	case "Do", "RoundTrip":
		return true
	default:
		return false
	}
}

func TestHTTPExecutionMethodInventoryIncludesTransportRoundTrip(t *testing.T) {
	for _, method := range []string{"Do", "RoundTrip"} {
		if !isHTTPExecutionMethod(method) {
			t.Errorf("HTTP execution method %q was not included in the inventory", method)
		}
	}
	if isHTTPExecutionMethod("Close") {
		t.Fatal("non-request method was included in the HTTP execution inventory")
	}
}

func expressionString(t *testing.T, fset *token.FileSet, node ast.Node) string {
	t.Helper()
	var output bytes.Buffer
	if err := printer.Fprint(&output, fset, node); err != nil {
		t.Fatalf("format Go expression: %v", err)
	}
	return output.String()
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}
