package normalizer

import (
	"github.com/bblfsh/python-driver/driver/normalizer/pyast"
	. "github.com/bblfsh/sdk/uast"
	. "github.com/bblfsh/sdk/uast/ann"
)
/*
Some stuff is missing from the current UAST spec to fully represent a Python AST. Issue:

https://github.com/bblfsh/documentation/issues/13

For a description of Python AST nodes:

https://greentreesnakes.readthedocs.io/en/latest/nodes.html

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
var AnnotationRules = On(Any).Self(
	On(Not(HasInternalType(pyast.Module))).Error("root must be Module"),
	On(HasInternalType(pyast.Module)).Roles(File).Descendants(

		On(HasInternalType(pyast.Expression)).Roles(Expression),
		On(HasInternalType(pyast.Expr)).Roles(Expression),
		On(HasInternalType(pyast.Str)).Roles(StringLiteral),

		On(HasInternalType(pyast.Call)).Roles(MethodInvocation).Children(
			On(HasInternalRole("args")).Children(On(Any).Roles(MethodInvocationArgument)),
			On(HasInternalRole("func")).Self(On(HasInternalRole("id"))).Roles(MethodInvocationName),
			On(HasInternalRole("func")).Self(On(HasInternalRole("attr"))).Roles(MethodInvocationName),
			On(HasInternalRole("func")).Self(On(HasInternalType(pyast.Attribute))).Children(
				On(HasInternalRole("id")).Roles(MethodInvocationObject),
			),
		),

		// Comments and non significative whitespace
		On(HasInternalRole(pyast.NoopLine)).Roles(Comment),
		On(HasInternalType(pyast.PreviousNoops)).Roles(Whitespace),
		On(HasInternalType(pyast.SameLineNoops)).Roles(Comment),
		//On(HasInternalType(pyast.RemainderNoops)).Roles(Whitespace),

		//On(HasInternalType(pyast.PreviousNoops)).Roles(Whitespace).Children(
		//	On(HasInternalRole(pyast.NoopLine)).Roles(Comment),
		//),
		On(HasInternalType(pyast.RemainderNoops)).Roles(Whitespace).Children(
			On(HasInternalRole(pyast.NoopLine)).Roles(Comment),
		),
		//On(HasInternalType(pyast.SameLineNoops)).Roles(Comment),

		// XXX integration tests cover until the above roles

		On(HasInternalType(pyast.Name)).Roles(SimpleIdentifier),
		On(HasInternalType(pyast.Assert)).Roles(Assert),

		On(HasInternalType(pyast.Constant)).Roles(Literal),
		On(HasInternalType(pyast.StringLiteral)).Roles(StringLiteral),
		// FIXME: should we make a distinction between StringLiteral and ByteLiteral on the UAST?
		On(HasInternalType(pyast.ByteLiteral)).Roles(StringLiteral),
		// FIXME: JoinedStr are the fstrings (f"my name is {name}"), they have a composite AST
		// with a body that is a list of StringLiteral + FormattedValue(value, conversion, format_spec)
		On(HasInternalType(pyast.JoinedStr)).Roles(StringLiteral),
		On(HasInternalType(pyast.NoneLiteral)).Roles(NullLiteral),
		On(HasInternalType(pyast.NumLiteral)).Roles(NumberLiteral),
		// FIXME: change these to ContainerLiteral/CompoundLiteral/whatever if they're added
		On(HasInternalType(pyast.Set)).Roles(Literal),
		On(HasInternalType(pyast.List)).Roles(Literal),
		On(HasInternalType(pyast.Dict)).Roles(Literal),
		On(HasInternalType(pyast.Tuple)).Roles(Literal),
		On(HasInternalType(pyast.Try)).Roles(Try).Children(
			On(HasInternalRole("body")).Roles(TryBody),
			On(HasInternalRole("finalbody")).Roles(TryFinally),
			// TODO: this is really a list, use descendents and search for ExceptHandlers?
			On(HasInternalRole("handlers")).Roles(TryCatch),
		),
		// FIXME: add OnPath Try.body (uast_type=ExceptHandler) => TryBody
		On(HasInternalType(pyast.TryExcept)).Roles(TryCatch),
		On(HasInternalType(pyast.TryFinally)).Roles(TryFinally),
		On(HasInternalType(pyast.Raise)).Roles(Throw),
		// FIXME: review, add path for the body and items childs
		// FIXME: withitem on Python to RAII on a resource and can aditionally create and alias on it,
		// both of which currently doesn't have representation in the UAST
		On(HasInternalType(pyast.With)).Roles(BlockScope),
		On(HasInternalType(pyast.Return)).Roles(Return),
		On(HasInternalType(pyast.Break)).Roles(Break),
		On(HasInternalType(pyast.Continue)).Roles(Continue),
		// FIXME: extract the test, orelse and the body to test-> IfCondition, orelse -> IfElse, body -> IfBody
		// UAST are first level members
		On(HasInternalType(pyast.If)).Roles(If),
		// One liner if, like a normal If but it will be inside an Assign (like the ternary if in C)
		// also applies the comment about the If
		On(HasInternalType(pyast.IfExp)).Roles(If),
		// FIXME: Import and ImportFrom can make an alias (name -> asname), extract it and put it as
		// uast.ImportAlias
		On(HasInternalType(pyast.Import)).Roles(ImportDeclaration),
		On(HasInternalType(pyast.ImportFrom)).Roles(ImportDeclaration),
		On(HasInternalType(pyast.ClassDef)).Roles(TypeDeclaration),
		// FIXME: add .args[].arg, .body, .name, .decorator_list[]
		On(HasInternalType(pyast.FunctionDef)).Roles(FunctionDeclaration),
		// FIXME: Internal keys for the ForEach: iter -> ?, target -> ?, body -> ForBody,
		//
		//	For => Foreach:
		//		body => ForBody
		//		iter => ForIter
		//		target => ForTarget
		//
		On(HasInternalType(pyast.For)).Roles(ForEach).Children(
			On(HasInternalRole("body")).Roles(ForBody),
			On(HasInternalRole("iter")).Roles(ForExpression),
			On(HasInternalRole("target")).Roles(ForUpdate),
		),
		// FIXME: while internal keys: body -> WhileBody, orelse -> ?, test -> WhileCondition
		On(HasInternalType(pyast.While)).Roles(While).Children(
			On(HasInternalRole("body")).Roles(WhileBody),
			On(HasInternalRole("test")).Roles(WhileCondition),

		),
		// FIXME: detect qualified 'Call.func' with a "Call.func.value" member and
		// "Call.func.ast_type" == attr (module/object calls) and convert the to this UAST:
		// MethodInvocation + MethodInvocationObject (func.value.id) + MethodInvocationName (func.attr)
		On(HasInternalType(pyast.Pass)).Roles(Noop),
		On(HasInternalType(pyast.Num)).Roles(NumberLiteral),
		//
		//	Assign => Assigment:
		//		targets[] => AssignmentVariable
		//		value     => AssignmentValue
		//
		On(HasInternalType(pyast.Assign)).Roles(Assignment).Children(
			On(HasInternalRole("targets")).Children(
				On(Any).Self().Roles(AssignmentVariable),
				On(HasInternalRole("value")).Roles(AssignmentVariable),
			),
		),
		// FIXME: this is the annotated assignment (a: annotation = 3) not exactly Assignment
		// it also lacks AssignmentValue and AssignmentVariable (see how to add them)
		On(HasInternalType(pyast.AnnAssign)).Roles(Assignment),
		// FIXME: this is the a += 1 style assigment
		On(HasInternalType(pyast.AugAssign)).Roles(Assignment),
	),
)

