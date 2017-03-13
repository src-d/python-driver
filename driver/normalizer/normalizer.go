package normalizer

import (
	"github.com/juanjux/python-driver/driver/normalizer/pyast"
	"github.com/bblfsh/sdk/uast"
)

/*
A lot of stuff is currently missing from the generated UAST. See:

https://github.com/bblfsh/documentation/issues/13

For a description of Python AST nodes:

https://greentreesnakes.readthedocs.io/en/latest/nodes.html?highlight=joinedstr#JoinedStr

	// Missing:
	GeneratorExp
	comprehension
	DictComp
	ListComp
	SetComp
	Yield
	YieldFrom
	AsyncFor
	AsyncFunctionDef
	AsyncWith => these three can be avoided and stored as For/FunctionDef/With if the save they
	             "async" keyword node
	Delete
	Call
	Lambda
	arguments
	arg              => arguments.args[list].arg (is both ast type and a key 'arg' pointing to the name)

	// Operators:
	Compare          => (comparators) .ops[list] = Eq | NotEq | Lt | LtE | Gt | GtE | Is | IsNot | In | NotIn
	BoolOp           => .boolop = And | Or
	BinOp            => .op = Add | Sub | Mult | MatMult | Div | Mod | Pow | LShift | RShift | BitOr |
	                          BitXor | BitAnd | FloorDiv
	UnaryOp          => .unaryop = Invert | Not | UAdd | USub

	// Other Keywords that probably could be SimpleIdentifier/Name subnodes in a parent "Keyword" AST node:
	Exec (body, globals, locals)
	Repr (value)
	Ellipsis ("..." for multidimensional arrays)
	Global
	Nonlocal
	Async
	Await
	Print

	// Other:
	Starred          => *expanded_list, could be translated to UnaryOp.Star
 */

// AnnotationRules for Python UAST.
var AnnotationRules uast.Rule = uast.Rules(
	uast.OnInternalType(pyast.Module).Role(uast.File),
	// FIXME: check how to add annotations and add them
	uast.OnInternalType(pyast.Name).Role(uast.SimpleIdentifier),
	uast.OnInternalType(pyast.Expression).Role(uast.File),
	uast.OnInternalType(pyast.Expr).Role(uast.File),
	uast.OnInternalType(pyast.expr).Role(uast.File),
	uast.OnInternalType(pyast.Assert).Role(uast.Assert),

	uast.OnInternalType(pyast.Constant).Role(uast.Literal),
	uast.OnInternalType(pyast.StringLiteral).Role(uast.StringLiteral),
	// FIXME: should we make a distinction between StringLiteral and ByteLiteral on the UAST?
	uast.OnInternalType(pyast.ByteLiteral).Role(uast.StringLiteral),
	// FIXME: JoinedStr are the fstrings (f"my name is {name}"), they have a composite AST
	// with a body that is a list of StringLiteral + FormattedValue(value, conversion, format_spec)
	uast.OnInternalType(pyast.JoinedStr).Role(uast.StringLiteral),
	uast.OnInternalType(pyast.NoneLiteral).Role(uast.NullLiteral),
	uast.OnInternalType(pyast.NumLiteral).Role(uast.NumberLiteral),
	// FIXME: change these to ContainerLiteral/CompoundLiteral/whatever if they're added
	uast.OnInternalType(pyast.Set).Role(uast.Literal),
	uast.OnInternalType(pyast.List).Role(uast.Literal),
	uast.OnInternalType(pyast.Dict).Role(uast.Literal),
	uast.OnInternalType(pyast.Tuple).Role(uast.Literal),
	uast.OnInternalType(pyast.Try).Role(uast.Try),
	// FIXME: add OnPath Try.body (uast_type=ExceptHandler) => TryBody
	uast.OnInternalType(pyast.TryExcept).Role(uast.TryCatch),
	uast.OnInternalType(pyast.TryFinally).Role(uast.TryFinally),
	uast.OnInternalType(pyast.Raise).Role(uast.Throw),
	// FIXME: review, add path for the body and items childs
	// FIXME: withitem on Python to RAII on a resource and can aditionally create and alias on it,
	// both of which currently doesn't have representation in the UAST
	uast.OnInternalType(pyast.With).Role(uast.BlockScope),
	uast.OnInternalType(pyast.Return).Role(uast.Return),
	uast.OnInternalType(pyast.Break).Role(uast.Break),
	uast.OnInternalType(pyast.Continue).Role(uast.Continue),
	// FIXME: extract the test, orelse and the body to test-> IfCondition, orelse -> IfElse, body -> IfBody
	// UAST are first level members
	uast.OnInternalType(pyast.If).Role(uast.If),
	// One liner if, like a normal If but it will be inside an Assign (like the ternary if in C)
	// also applies the comment about the If
	uast.OnInternalType(pyast.IfExp).Role(uast.If),
	// FIXME: Import and ImportFrom can make an alias (name -> asname), extract it and put it as
	// uast.ImportAlias
	uast.OnInternalType(pyast.Import).Role(uast.Import),
	uast.OnInternalType(pyast.ImportFrom).Role(uast.Import),
	uast.OnInternalType(pyast.ClassDef).Role(uast.TypeDeclaration),
	// FIXME: add .args[].arg, .body, .name, .decorator_list[]
	uast.OnInternalType(pyast.FunctionDef).Role(uast.FunctionDeclaration),
	// FIXME: Internal keys for the ForEach: iter -> ?, target -> ?, body -> ForBody,
	uast.OnInternalType(pyast.For).Role(uast.ForEach),
	// FIXME: while internal keys: body -> WhileBody, orelse -> ?, test -> WhileCondition
	uast.OnInternalType(pyast.While).Role(uast.While),
	// FIXME: detect qualified 'Call.func' with a "Call.func.value" member and
	// "Call.func.ast_type" == attr (module/object calls) and convert the to this UAST:
	// MethodInvocation + MethodInvocationObject (func.value.id) + MethodInvocationName (func.attr)
	uast.OnInternalType(pyast.Pass).Role(uast.Noop),
	uast.OnInternalType(pyast.Str).Role(uast.StringLiteral),
	uast.OnInternalType(pyast.Num).Role(uast.NumberLiteral),
	uast.OnInternalType(pyast.Assign).Role(uast.Assignment),
	// FIXME: this is the annotated assignment (a: annotation = 3) not exactly Assignment
	// it also lacks AssignmentValue and AssignmentVariable (see how to add them)
	uast.OnInternalType(pyast.AnnAssign).Role(uast.Assignment),
	// FIXME: this is the a += 1 style assigment
	uast.OnInternalType(pyast.AugAssign).Role(uast.Assignment),
)

// Annotate annotates the given Java UAST.
func Annotate(n *uast.Node) error {
	return uast.PreOrderVisit(n, AnnotationRules)
}

