package normalizer

import (
	"github.com/bblfsh/sdk/uast"
)

var NativeToNoder = &uast.BaseToNoder{
	InternalTypeKey: "internalClass",
	LineKey:         "line",
	OffsetKey:       "startPosition",
	//TODO: Should this be part of the UAST rules?
	TokenKeys: map[string]bool{
		"identifier":        true, // SimpleName
		"escapedValue":      true, // StringLiteral
		"keyword":           true, // Modifier
		"primitiveTypeCode": true, // ?
	},
	SyntheticTokens: map[string]string{
		"PackageDeclaration": "package",
		"IfStatement":        "if",
		"NullLiteral":        "null",
	},
	//TODO: add names of children (e.g. elseStatement) as
	//      children node properties.
}
