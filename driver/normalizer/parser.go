package normalizer

import (
	"github.com/bblfsh/sdk/protocol/driver"
	"github.com/bblfsh/sdk/protocol/native"
)

var ToNoder = &native.ObjectToNoder{
	InternalTypeKey: "ast_type",
	LineKey:         "lineno",
	EndLineKey:      "end_lineno",
	ColumnKey:       "col_offset",
	EndColumnKey:    "end_col_offset",
	PositionFill:    native.OffsetFromLineCol,

	TokenKeys: map[string]bool{
		"name": true,
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
		"Mod":       "%%",
		"Pow":       "**",
		"AugAssign": "?=",
		"BitAnd":    "&",
		"BitOr":     "|",
		"BitXor":    "^",
		"LShift":    "<<",
		"RShift":    ">>",
		"Eq":        "==",
		"NotEq":     "!=",
		"Not":       "!",
		"Lt":        "<",
		"LtE":       "<=",
		"Gt":        ">",
		"GtE":       ">=",
		"Is":        "is",
		"IsNot":     "not is",
		"In":        "in",
		"NotIn":     "not in",
		"UAdd":      "+",
		"USub":      "-",
		"Invert":    "~",
		"Pass":      "pass",
	},
	PromoteAllPropertyLists: false,
	PromotedPropertyLists: map[string]map[string]bool{
		"If":       {"body": true, "orelse": true},
		"For":      {"body": true, "orelse": true},
		"AsyncFor": {"body": true, "orelse": true},
		"While":    {"body": true, "orelse": true},
		"Compare":  {"comparators": true, "ops": true},
		// FIXME: check call.keywords
		//"Call"                    : { "args": true, "keywords": true},
		"With":        {"body": true, "items": true},
		"FunctionDef": {"body": true, "decorator_list": true},
		"Lambda":      {"body": true},
		"arguments":   {"defaults": true},
		"Try":         {"body": true, "orelse": true, "finalbody": true},
		"Raise":       {"args": true},
		"ClassDef":    {"body": true, "bases": true, "decorator_list": true, "keywords": true},
	},
	PromotedPropertyStrings: map[string]map[string]bool{
		"alias":         {"asname": true},
		"ImportFrom":    {"module": true},
		"ExceptHandler": {"name": true},
	},
	// FIXME: test[ast_type=Compare].comparators is a list?? (should be "right")
}

// ParserBuilder creates a parser that transform source code files into *uast.Node.
func ParserBuilder(opts driver.ParserOptions) (parser driver.Parser, err error) {
	parser, err = native.ExecParser(ToNoder, opts.NativeBin)
	if err != nil {
		return
	}

	switch ToNoder.PositionFill {
	case native.OffsetFromLineCol:
		parser = &driver.TransformationParser{
			Parser:         parser,
			Transformation: driver.FillOffsetFromLineCol,
		}
	case native.LineColFromOffset:
		parser = &driver.TransformationParser{
			Parser:         parser,
			Transformation: driver.FillLineColFromOffset,
		}
	}

	return
}
