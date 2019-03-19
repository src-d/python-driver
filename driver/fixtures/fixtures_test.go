package fixtures

import (
	"path/filepath"
	"testing"

	"github.com/bblfsh/python-driver/driver/normalizer"
	"gopkg.in/bblfsh/sdk.v2/driver"
	"gopkg.in/bblfsh/sdk.v2/driver/fixtures"
	"gopkg.in/bblfsh/sdk.v2/driver/native"
)

const projectRoot = "../../"

var Suite = &fixtures.Suite{
	Lang: "python",
	Ext:  ".py",
	Path: filepath.Join(projectRoot, fixtures.Dir),
	NewDriver: func() driver.Native {
		return native.NewDriverAt(filepath.Join(projectRoot, "build/bin/native"), native.UTF8)
	},
	Transforms: normalizer.Transforms,
	BenchName:  "issue_server101",
	Semantic: fixtures.SemanticConfig{
		BlacklistTypes: []string{
			"AsyncFunctionDef",
			"Attribute",
			"BoolLiteral",
			"Bytes",
			"FunctionDef",
			"ImportFrom",
			"Name",
			"NoopLine",
			"NoopSameLine",
			"Str",
			"StringLiteral",
			"alias",
			"arg",
			"kwarg",
			"kwonly_arg",
			"vararg",
		},
	},
}

func TestPythonDriver(t *testing.T) {
	Suite.RunTests(t)
}

func BenchmarkPythonDriver(b *testing.B) {
	Suite.RunBenchmarks(b)
}
