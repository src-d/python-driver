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

	// Missing: =======================================
	// Comprehensions
	comprehension: DictComp ListComp SetComp

	Ellipsis ("..." for multidimensional arrays)

	Lambda (wait for FunctionDef to be complete)

	arg => arguments.args[list].arg (is both ast type and a key 'arg' pointing to the name)
			 => for python we would need:
			 	- FunctionDefArg (a)
			 	- FunctionDefArgDefaultValue (a = 3)
			 	- FunctionDefArgAnnotation (a: int)
			 	- FunctionDefVarArgsList (*args)
			 	- FunctionDefVarArgsMap (**kwargs)
	// TODO: make a full proposal for FunctionDef

	// To do in rules (TODO): ========================================
	exec, repr, print: are nodes in the Python 2 AST but they take the form of a functioncall.
	Convert them to function calls using the rules
	Convert ellipsis to a SimpleIdentifier with the name "pyellipsis".

	// Merged or added as issue: ==============================================
	Issue #52: Global Nonlocal
	Issue #53: Async, Await, AsyncFor AsyncFunctionDef AsyncWith => these three can be avoided and stored as
		For/FunctionDef/With if the save they "async" keyword node

	// Operators:
	PR: Compare          => (comparators) .ops[list] = Eq | NotEq | Lt | LtE | Gt | GtE | Is | IsNot | In | NotIn
	PR: BoolOp           => .boolop = And | Or
	PR: BinOp            => .op = Add | Sub | Mult | MatMult | Div | Mod | Pow | LShift | RShift | BitOr |
	                          BitXor | BitAnd | FloorDiv
	PR: UnaryOp          => .unaryop = Invert | Not | UAdd | USub
	PR #55: Delete
	PR #56: Yield and YieldFrom
	PR #57	ListExpansion (Starred) MapExpansion (**)
	PR #58: BlockResource (the "thing" in "with thing:"), content_expr in the Python AST
		>>> c = "with thing as t: t"
		>>> ast.dump(ast.parse(c))
		"Module(body=[With(
				items=[
				   withitem(context_expr=Name(id='thing', ctx=Load()), optional_vars=Name(id='t', ctx=Store()))
				  ],
				  body=[
				    Expr(value=Name(id='t', ctx=Load())
				    ) ] )
			       ]
			  )"
    PR #59:
		Subscript ->
			a[1] -> Index value=NumLiteral
			a[1:2] -> Slice lower=NumLiteral upper=NumLiteral
			a[1:2,3:4] -> ExtSlice -> [Slice, Slice]

	// Noop (decently parsed by the rules even if unorthodox):
	"else" clauses for for/while/try -> Added as "IfElse" child nodes of the for/while/try
	"stride" third element in slices

 */

 // TODO: all the "orelse" subnodes of for/while/try as "IfElse" childs
 // TODO: add the "stride", third element of slices in some way once we've the index/slice roles
var AnnotationRules = On(Any).Self(
	On(Not(HasInternalType(pyast.Module))).Error("root must be Module"),
	On(HasInternalType(pyast.Module)).Roles(File).Descendants(
		// Comparison operators
		On(HasInternalType(pyast.Eq).Roles(OpEqual),
		On(HasInternalType(pyast.NotEq).Roles(OpNotEqual),
		On(HasInternalType(pyast.Lt).Roles(OpLessThan),
		On(HasInternalType(pyast.LtE).Roles(OpLessThanEqual),
		On(HasInternalType(pyast.Gt).Roles(OpGreaterThan),
		On(HasInternalType(pyast.GtE).Roles(OpGreaterThanEqual),
		On(HasInternalType(pyast.Is).Roles(OpSame),
		On(HasInternalType(pyast.IsNot).Roles(OpNotSame),
		On(HasInternalType(pyast.In).Roles(OpContains),
		On(HasInternalType(pyast.NotIn).Roles(OpNotContains),

		// FIXME: boolliteral should probably be added to the UAST
		On(HasInternalType(pyast.StringLiteral)).Roles(StringLiteral),
		On(HasInternalType(pyast.ByteLiteral)).Roles(ByteStringLiteral),
		On(HasInternalType(pyast.NumLiteral)).Roles(NumberLiteral),
		On(HasInternalType(pyast.Str)).Roles(StringLiteral),
		On(HasInternalType(pyast.BoolLiteral)).Roles(BooleanLiteral),
		// FIXME: JoinedStr are the fstrings (f"my name is {name}"), they have a composite AST
		// with a body that is a list of StringLiteral + FormattedValue(value, conversion, format_spec)
		On(HasInternalType(pyast.JoinedStr)).Roles(StringLiteral),
		On(HasInternalType(pyast.NoneLiteral)).Roles(NullLiteral),
		// FIXME: change these to ContainerLiteral/CompoundLiteral/whatever if they're added
		On(HasInternalType(pyast.Set)).Roles(SetLiteral),
		On(HasInternalType(pyast.List)).Roles(ListLiteral),
		On(HasInternalType(pyast.Dict)).Roles(MapLiteral),
		On(HasInternalType(pyast.Tuple)).Roles(TupleLiteral),

		// FIXME: add .args[].arg, .body, .name, .decorator_list[]
		// FIXME XXX BUG: Call is for calls, not for definitions!
		On(HasInternalType(pyast.FunctionDef)).Roles(FunctionDeclaration),
		On(HasInternalType(pyast.Call)).Roles(Call).Children(
			On(HasInternalRole("args")).Roles(CallPositionalArgument),
			On(HasInternalRole("func")).Self(On(HasInternalRole("id"))).Roles(CallCallee),
			On(HasInternalRole("func")).Self(On(HasInternalRole("attr"))).Roles(CallCallee),
			On(HasInternalRole("func")).Self(On(HasInternalType(pyast.Attribute))).Children(
				On(HasInternalRole("id")).Roles(CallReceiver),
			),
		),

		//
		//	Assign => Assigment:
		//		targets[] => AssignmentVariable
		//		value     => AssignmentValue
		//
		On(HasInternalType(pyast.Assign)).Roles(Assignment).Children(
			On(HasInternalRole("targets")).Roles(AssignmentVariable),
			On(HasInternalRole("value")).Roles(AssignmentValue),
		),

		On(HasInternalType(pyast.Expression)).Roles(Expression),
		On(HasInternalType(pyast.Expr)).Roles(Expression),
		On(HasInternalType(pyast.Name)).Roles(SimpleIdentifier),

		// Comments and non significative whitespace
		On(HasInternalType(pyast.SameLineNoops)).Roles(Comment),
		On(HasInternalType(pyast.PreviousNoops)).Roles(Whitespace).Children(
			On(HasInternalRole("lines")).Roles(Comment),
		),
		On(HasInternalType(pyast.RemainderNoops)).Roles(Whitespace).Children(
			On(HasInternalRole("lines")).Roles(Comment),
		),

		On(HasInternalType(pyast.Constant)).Roles(Literal),
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
		On(HasInternalType(pyast.If)).Roles(If).Children(
			On(HasInternalRole("body")).Roles(IfBody),
			On(HasInternalRole("orelse")).Roles(IfElse),
			On(HasInternalType(pyast.Compare)).Roles(IfCondition),
		),
		// One liner if, like a normal If but it will be inside an Assign (like the ternary if in C)
		// also applies the comment about the If
		On(HasInternalType(pyast.IfExp)).Roles(If),
		// FIXME: Import and ImportFrom can make an alias (name -> asname), extract it and put it as
		// uast.ImportAlias
		On(HasInternalType(pyast.Import)).Roles(ImportDeclaration),
		On(HasInternalType(pyast.ImportFrom)).Roles(ImportDeclaration),
		On(HasInternalType(pyast.Alias)).Roles(ImportPath),
		On(HasInternalType(pyast.ClassDef)).Roles(TypeDeclaration),
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
		On(HasInternalType(pyast.Pass)).Roles(Noop),
		On(HasInternalType(pyast.Num)).Roles(NumberLiteral),
		// FIXME: this is the annotated assignment (a: annotation = 3) not exactly Assignment
		// it also lacks AssignmentValue and AssignmentVariable (see how to add them)
		On(HasInternalType(pyast.AnnAssign)).Roles(Assignment),
		// FIXME: this is the a += 1 style assigment
		On(HasInternalType(pyast.AugAssign)).Roles(Assignment),
		On(HasInternalType(pyast.Assert)).Roles(Assert),
	),
)

