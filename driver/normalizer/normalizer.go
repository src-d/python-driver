package normalizer

import (
	"github.com/juanjux/python-driver/driver/normalizer/pyast"

	"github.com/bblfsh/sdk/uast"
)

// NewToNoder creates a new uast.ToNoder to convert
// Python ASTs to UAST.
func NewToNoder() uast.ToNoder {
	return &uast.BaseToNoder{
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
}

// AnnotationRules for Python UAST.
var AnnotationRules uast.Rule = uast.Rules(
	uast.OnInternalType(pyast.CompilationUnit).Role(uast.File),
	uast.OnInternalType(pyast.PackageDeclaration).Role(uast.PackageDeclaration),
	uast.OnInternalType(pyast.MethodDeclaration).Role(uast.FunctionDeclaration),
	uast.OnInternalType(pyast.ImportDeclaration).Role(uast.ImportDeclaration),
	uast.OnInternalType(pyast.TypeDeclaration).Role(uast.TypeDeclaration),
	uast.OnInternalType(pyast.ImportDeclaration, pyast.QualifiedName).Role(uast.ImportPath),
	uast.OnInternalType(pyast.QualifiedName).Role(uast.QualifiedIdentifier),
	uast.OnInternalType(pyast.SimpleName).Role(uast.SimpleIdentifier),
	uast.OnInternalType(pyast.Block).Role(uast.BlockScope, uast.Block),
	uast.OnInternalType(pyast.ExpressionStatement).Role(uast.Statement),
	uast.OnInternalType(pyast.ReturnStatement).Role(uast.Return, uast.Statement),
	uast.OnInternalType(pyast.MethodInvocation).Role(uast.MethodInvocation),
	uast.OnInternalType(pyast.IfStatement).Role(uast.If, uast.Statement),
	uast.OnInternalRole("elseStatement").Role(uast.IfElse, uast.Statement),
	uast.OnPath(uast.OnInternalType(pyast.Assignment)).Role(uast.Assignment),
	uast.OnPath(uast.OnInternalType(pyast.Assignment), uast.OnInternalRole("leftHandSide")).Role(uast.AssignmentVariable),
	uast.OnPath(uast.OnInternalType(pyast.Assignment), uast.OnInternalRole("rightHandSide")).Role(uast.AssignmentValue),
	//TODO: IfBody, IfCondition
	uast.OnInternalType(pyast.NullLiteral).Role(uast.NullLiteral, uast.Literal),
	uast.OnInternalType(pyast.StringLiteral).Role(uast.StringLiteral, uast.Literal),
	uast.OnInternalType(pyast.NumberLiteral).Role(uast.NumberLiteral, uast.Literal),
	uast.OnInternalType(pyast.TypeLiteral).Role(uast.TypeLiteral, uast.Literal),
	uast.OnInternalType(pyast.ThisExpression).Role(uast.This, uast.Expression),
	//TODO: synchronized
	//TODO: try-with-resources
	uast.OnInternalType(pyast.Javadoc).Role(uast.Documentation, uast.Comment),
)

// Annotate annotates the given Java UAST.
func Annotate(n *uast.Node) error {
	return uast.PreOrderVisit(n, AnnotationRules)
}

