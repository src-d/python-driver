package normalizer

import (
	"github.com/bblfsh/sdk/protocol/driver"
	"github.com/bblfsh/sdk/protocol/native"
)

var ToNoder = &native.ObjectToNoder{
	InternalTypeKey: "ast_type",
	LineKey:         "lineno",
	ColumnKey:       "col_offset",

	TokenKeys: map[string]bool{
		"module": true, // Module on ImportFrom
		"name":   true,
		//"asname":          true, // Alias from ImportFrom
		"id":           true, // Name nodes
		"attr":         true, // something.attr
		"arg":          true, // function arguments
		"LiteralValue": true, // string/num/byte/constant literal
		"noop_line":    true, // Comment/Noop (non significative whitespace)
	},
	SyntheticTokens: map[string]string{
		"Print":     "print",
		"Ellipsis":  "PythonEllipsisOperator",
		"Add":       "+",
		"Sub":       "-",
		"Mult":      "*",
		"Div":       "/",
		"FloorDiv":  "//",
		"Mod":       "%",
		"Pow":       "^",
		"AugAssign": "?=",
	},
	PromoteAllPropertyLists: false,
	PromotedPropertyLists: map[string]map[string]bool{
		"If":    {"body": true, "orelse": true},
		"For":   {"body": true, "orelse": true},
		"While": {"body": true, "orelse": true},
		// FIXME: check if promotion is needed in this case
		"Compare":    {"comparators": true, "ops": true},
		"Import":     {"names": true},
		"ImportFrom": {"names": true},
		// FIXME: check call.keywords
		//"Call"                    : { "args": true, "keywords": true},
		"With":                    {"body": true, "items": true},
		"FunctionDef":             {"body": true, "decorator_list": true},
		"arguments":               {"defaults": true},
		"Try":                     {"body": true, "orelse": true, "finalbody": true, "handlers": true},
		"Raise":                   {"args": true},
		"ClassDef":                {"body": true, "bases": true, "decorator_list": true, "keywords": true},
		"ListComp":                {"generators": true},
		"ListComp.generators":     {"ifs": true},
		"ListComp.generators.ifs": {"comparators": true, "ops": true},
	},
	// FIXME: test[ast_type=Compare].comparators is a list?? (should be "right")
}

// ASTParserBuilder creates a parser that transform source code files into *uast.Node.
func ASTParserBuilder(opts driver.ASTParserOptions) (driver.ASTParser, error) {
	parser, err := native.ExecParser(ToNoder, opts.NativeBin)
	if err != nil {
		return nil, err
	}

	return parser, nil
}
