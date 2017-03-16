package normalizer

import (
	"github.com/bblfsh/sdk/uast"
)

var NativeToNoder = &uast.BaseToNoder{
	// FIXME: ColumnKey? (col_offset in Python)
	InternalTypeKey: "ast_type",
	LineKey:         "lineno",

	// FIXME: if ones matches several? (ImportFrom -> Names can have name and asname)
	TokenKeys: map[string]bool{
		"module":            true, // Module on ImportFrom
		"name":              true,
		//"asname":            true, // Alias from ImportFrom
		"id":                true, // Name nodes
		"attr":              true, // something.attr
		"arg":               true, // function arguments
	},
	SyntheticTokens: map[string]string{
		"NullLiteral":        "null",
	},
}
