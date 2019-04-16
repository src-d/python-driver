package normalizer

import (
	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/role"
	. "github.com/bblfsh/sdk/v3/uast/transformer"
)

var Native = Transformers([][]Transformer{
	{
		ResponseMetadata{
			TopLevelIsRootNode: false,
		},
	},
	{Mappings(Annotations...)},
	{RolesDedup()},
}...)

func annotateTypeToken(typ, token string, roles ...role.Role) Mapping {
	return AnnotateType(typ,
		FieldRoles{
			uast.KeyToken: {Add: true, Op: String(token)},
		}, roles...)
}

var funcBodyRoles = Roles(role.Function, role.Declaration, role.Body)
var funcDecoRoles = Roles(role.Function, role.Declaration, role.Incomplete)

func functionAnnotate(typ string, roles ...role.Role) Mapping {
	return AnnotateType(typ, MapObj(Obj{
		"decorator_list": Var("decors"),
		"body":           Var("body_stmts"),
		"name":           Var("name"),
	}, Obj{
		"decorator_list": Obj{
			uast.KeyType:  String("FunctionDef.decorators"),
			uast.KeyRoles: funcDecoRoles,
			"decorators":  Var("decors"),
		},
		"body": Obj{
			uast.KeyType:  String("FunctionDef.body"),
			uast.KeyRoles: funcBodyRoles,
			"body_stmts":  Var("body_stmts"),
		},
		uast.KeyToken: Var("name"),
	}), roles...)
}

func withAnnotate(typ string, roles ...role.Role) Mapping {
	return AnnotateType(typ, MapObj(Obj{
		"body":  Var("body_stmts"),
		"items": Var("itms"),
	}, Obj{
		"body": Obj{
			uast.KeyType:  String("With.body"),
			uast.KeyRoles: Roles(role.Block, role.Scope, role.Body, role.Incomplete),
			"body_stmts":  Var("body_stmts"),
		},
		"items": Obj{
			uast.KeyType:  String("With.items"),
			uast.KeyRoles: Roles(role.Block, role.Scope, role.Incomplete),
			"items":       Var("itms"),
		},
	}), role.Block, role.Scope, role.Statement)
}

func loopAnnotate(typ string, mainRole role.Role, roles ...role.Role) Mapping {
	return AnnotateType(typ, MapObj(Obj{
		"body":   Var("body_stmts"),
		"orelse": Var("else_stmts"),
	}, Obj{
		"body": Obj{
			uast.KeyType:  String("For.body"),
			uast.KeyRoles: Roles(mainRole, role.Body),
			"body_stmts":  Var("body_stmts"),
		},
		"orelse": Obj{
			uast.KeyType:  String("For.orelse"),
			uast.KeyRoles: Roles(mainRole, role.Body, role.Else),
			"else_stmts":  Var("else_stmts"),
			uast.KeyToken: String("else"),
		},
	}), roles...)
}

var Annotations = []Mapping{
	// FIXME: doesnt work
	AnnotateType("Module", nil, role.File, role.Module),

	// Comparison operators
	// in Python, with internaltype)
	annotateTypeToken("Eq", "==", role.Operator, role.Relational, role.Equal),
	annotateTypeToken("NotEq", "!=", role.Operator, role.Relational, role.Not, role.Equal),
	annotateTypeToken("Lt", "<", role.Operator, role.Relational, role.LessThan),
	annotateTypeToken("LtE", "<=", role.Operator, role.Relational, role.LessThanOrEqual),
	annotateTypeToken("Gt", ">", role.Operator, role.Relational, role.GreaterThan),
	annotateTypeToken("GtE", ">=", role.Operator, role.Relational, role.GreaterThanOrEqual),
	annotateTypeToken("Is", "is", role.Operator, role.Relational, role.Identical),
	annotateTypeToken("IsNot", "is not", role.Operator, role.Relational, role.Not, role.Identical),
	annotateTypeToken("In", "in", role.Operator, role.Relational, role.Contains),
	annotateTypeToken("NotIn", "not in", role.Operator, role.Relational, role.Not, role.Contains),

	// Arithmetic operators
	annotateTypeToken("Add", "+", role.Operator, role.Arithmetic, role.Add),
	annotateTypeToken("Sub", "-", role.Operator, role.Arithmetic, role.Substract),
	annotateTypeToken("Mult", "*", role.Operator, role.Arithmetic, role.Multiply),
	annotateTypeToken("MatMult", "@", role.Operator, role.Arithmetic, role.Multiply, role.Incomplete),
	annotateTypeToken("Div", "/", role.Operator, role.Arithmetic, role.Divide),
	annotateTypeToken("Mod", "%", role.Operator, role.Arithmetic, role.Module),
	annotateTypeToken("FloorDiv", "//", role.Operator, role.Arithmetic, role.Divide, role.Incomplete),
	annotateTypeToken("Pow", "**", role.Operator, role.Arithmetic, role.Incomplete),

	// Bitwise operators
	annotateTypeToken("LShift", "<<", role.Operator, role.Bitwise, role.LeftShift),
	annotateTypeToken("RShift", ">>", role.Operator, role.Bitwise, role.RightShift),
	annotateTypeToken("BitOr", "|", role.Operator, role.Bitwise, role.Or),
	annotateTypeToken("BitXor", "^", role.Operator, role.Bitwise, role.Xor),
	annotateTypeToken("BitAnd", "&", role.Operator, role.Bitwise, role.And),

	// Boolean operators
	// Not applying the "Binary" role since even while in the Python code
	// boolean operators use (seemingly binary) infix notation, the generated
	// AST nodes use prefix.
	annotateTypeToken("And", "and", role.Operator, role.Boolean, role.And),
	annotateTypeToken("Or", "or", role.Operator, role.Boolean, role.Or),
	annotateTypeToken("Not", "not", role.Operator, role.Boolean, role.Not),
	AnnotateType("UnaryOp", nil, role.Operator, role.Boolean, role.Unary, role.Expression),

	// Unary operators
	annotateTypeToken("Invert", "~", role.Operator, role.Unary, role.Bitwise, role.Not),
	annotateTypeToken("UAdd", "+", role.Operator, role.Unary, role.Bitwise, role.Positive),
	annotateTypeToken("USub", "-", role.Operator, role.Unary, role.Bitwise, role.Negative),

	// Compound Literals
	// another grouping node like "arguments"
	AnnotateType("Set", nil, role.Literal, role.Set, role.Expression, role.Primitive),
	AnnotateType("List", nil, role.Literal, role.List, role.Expression, role.Primitive),
	AnnotateType("Tuple", nil, role.Literal, role.Tuple, role.Expression, role.Primitive),

	// Expressions
	AnnotateType("Expression", nil, role.Expression),
	AnnotateType("Expr", nil, role.Expression),
	// grouping node for boolean expressions:
	AnnotateType("BoolOp", nil, role.Literal, role.Boolean, role.Incomplete),

	// Misc
	annotateTypeToken("Return", "return", role.Return, role.Statement),
	annotateTypeToken("Break", "break", role.Break, role.Statement),
	annotateTypeToken("Continue", "continue", role.Continue, role.Statement),
	// Python very odd ellipsis operator. Has a special rule in tonoder synthetic tokens
	// map to load it with the token "PythonEllipsisuast.Operator" and gets the role uast.Identifier
	annotateTypeToken("Ellipsis", "...", role.Identifier, role.Incomplete),
	annotateTypeToken("Delete", "del", role.Statement, role.Incomplete),
	annotateTypeToken("Await", "await", role.Statement, role.Incomplete),
	annotateTypeToken("Global", "global", role.Statement, role.Visibility, role.World, role.Incomplete),
	annotateTypeToken("Nonlocal", "nonlocal", role.Statement, role.Visibility, role.Module, role.Incomplete),
	annotateTypeToken("Yield", "yield", role.Statement, role.Return, role.Incomplete),
	annotateTypeToken("With", "with"),
	annotateTypeToken("For", "for"),
	annotateTypeToken("If", "if"),
	annotateTypeToken("Try", "try"),
	annotateTypeToken("While", "while"),
	annotateTypeToken("YieldFrom", "yield from", role.Statement, role.Return, role.Incomplete),
	AnnotateType("Subscript", nil, role.Expression, role.Incomplete),
	AnnotateType("Index", nil, role.Expression, role.Incomplete),
	// FIXME: no support for slices/ranges in the roles
	AnnotateType("Slice", nil, role.Expression, role.Incomplete),
	AnnotateType("ExtSlice", nil, role.Expression, role.Incomplete),
	annotateTypeToken("Pass", "pass", role.Noop, role.Statement),
	annotateTypeToken("Assert", "assert", role.Assert, role.Statement),

	AnnotateType("Name", FieldRoles{"id": {Rename: uast.KeyToken}},
		role.Identifier, role.Expression),
	AnnotateType("Attribute", FieldRoles{"attr": {Rename: uast.KeyToken}},
		role.Identifier, role.Expression),
	AnnotateType("QualifiedIdentifier", nil, role.Identifier, role.Expression, role.Qualified),

	// Binary Expressions
	AnnotateType("BinOp", ObjRoles{
		"left":  {role.Expression, role.Binary, role.Left},
		"right": {role.Expression, role.Binary, role.Right},
		"op":    {role.Binary},
	}, role.Expression, role.Binary),

	// Primitive Literals
	AnnotateType("Str", FieldRoles{"s": {Rename: uast.KeyToken}},
		role.Literal, role.String, role.Expression, role.Primitive),
	AnnotateType("Bytes", FieldRoles{"s": {Rename: uast.KeyToken}},
		role.Literal, role.ByteString, role.Expression, role.Primitive),
	AnnotateType("StringLiteral", FieldRoles{"s": {Rename: uast.KeyToken}},
		role.Literal, role.String, role.Expression, role.Primitive),
	AnnotateType("BoolLiteral", FieldRoles{"LiteralValue": {Rename: uast.KeyToken}},
		role.Literal, role.Boolean, role.Expression, role.Primitive),
	annotateTypeToken("NoneLiteral", "None", role.Literal, role.Null, role.Expression, role.Primitive),
	AnnotateType("Num", FieldRoles{"n": {Rename: uast.KeyToken}},
		role.Expression, role.Literal, role.Number, role.Primitive),
	AnnotateType("BoolLiteral", FieldRoles{"LiteralValue": {Rename: uast.KeyToken}},
		role.Expression, role.Literal, role.Boolean, role.Primitive),
	AnnotateType("Dict", FieldRoles{
		"keys":   {Arr: true, Roles: role.Roles{role.Map, role.Key}},
		"values": {Arr: true, Roles: role.Roles{role.Map, role.Value}},
	}, role.Expression, role.Literal, role.Primitive, role.Map),

	// another grouping node like "arguments"
	AnnotateType("JoinedStr", nil, role.Expression, role.Literal, role.Primitive, role.String, role.Incomplete),
	AnnotateType("FormattedValue", nil, role.Expression, role.Incomplete),

	//
	//	Assign => Assigment:
	//		targets[] => Left
	//		value	  => Right
	//
	AnnotateType("Assign", FieldRoles{
		"targets": {Arr: true, Roles: role.Roles{role.Left}},
		"value":   {Roles: role.Roles{role.Right}},
	}, role.Binary, role.Expression, role.Assignment),

	AnnotateType("AugAssign", ObjRoles{
		"op":     {role.Operator},
		"target": {role.Right},
		"value":  {role.Left},
	}, role.Binary, role.Expression, role.Operator, role.Assignment,
	),

	// Exceptions
	// Adds a parent node for each these properties with direct list values
	AnnotateType("Try", MapObj(Obj{
		"body":      Var("body_stmts"),
		"finalbody": Var("final_stmts"),
		"handlers":  Var("handlers_list"),
		"orelse":    Var("else_stmts"),
	}, Obj{
		"body": Obj{
			uast.KeyType:  String("Try.body"),
			uast.KeyRoles: Roles(role.Try, role.Body),
			"body_stmts":  Var("body_stmts"),
		},
		"finalbody": Obj{
			uast.KeyType:  String("Try.finalbody"),
			uast.KeyRoles: Roles(role.Try, role.Finally),
			"final_stmts": Var("final_stmts"),
			uast.KeyToken: String("finally"),
		},
		"handlers": Obj{
			uast.KeyType:  String("Try.handlers"),
			uast.KeyRoles: Roles(role.Try, role.Catch),
			"handlers":    Var("handlers_list"),
			uast.KeyToken: String("except"),
		},
		"orelse": Obj{
			uast.KeyType:  String("Try.else"),
			uast.KeyRoles: Roles(role.Try, role.Else),
			uast.KeyToken: String("else"),
		},
	}), role.Try, role.Statement),

	// python 2 exception handling
	AnnotateType("TryExcept", nil, role.Try, role.Catch, role.Statement),
	AnnotateType("ExceptHandler", FieldRoles{"name": {Rename: uast.KeyToken}},
		role.Try, role.Catch, role.Identifier),
	AnnotateType("TryFinally", nil, role.Try, role.Finally, role.Statement),
	AnnotateType("Raise", nil, role.Throw),

	AnnotateType("Raise",
		FieldRoles{
			"exc":         {Opt: true, Roles: role.Roles{role.Call}},
			uast.KeyToken: {Add: true, Op: String("raise")},
		}, role.Throw, role.Statement),

	// With
	withAnnotate("With"),
	withAnnotate("AsyncWith"),
	AnnotateType("withitem", nil, role.Identifier, role.Expression, role.Incomplete),

	// uast.List/uast.Map/uast.Set comprehensions. We map the "for x in y" to uast.For, uast.Iterator (foreach)
	// roles and the "if something" to uast.If* roles.
	// FIXME: missing the top comprehension roles in the UAST, change once they've been
	// merged
	AnnotateType("ListComp", nil, role.List, role.For, role.Expression),
	AnnotateType("DictComp", nil, role.Map, role.For, role.Expression),
	AnnotateType("SetComp", nil, role.Set, role.For, role.Expression),

	// FIXME: once we have an async Role we should interpret the is_async property
	AnnotateType("comprehension", FieldRoles{
		"ifs":    {Arr: true, Roles: role.Roles{role.If, role.Condition}},
		"iter":   {Roles: role.Roles{role.For, role.Update, role.Statement}},
		"target": {Roles: role.Roles{role.For, role.Expression}},
	}, role.For, role.Iterator, role.Expression, role.Incomplete),

	// Python annotations for variables, function argument or return values doesn't
	// have any semantic information by themselves and this we consider it comments
	// (some preprocessors or linters can use them, the runtimes ignore them). The
	// TOKEN will take the annotation in the UAST node so the information is keept in
	// any case.
	AnnotateType("AnnAssign", nil, role.Operator, role.Binary, role.Assignment),
	AnnotateType("annotation", nil, role.Annotation, role.Noop),
	AnnotateType("returns", nil, role.Annotation, role.Noop),

	// Function Declaratations
	functionAnnotate("FunctionDef", role.Function, role.Declaration, role.Name, role.Identifier),
	// FIXME: Incomplete for lacking of an Async role
	functionAnnotate("AsyncFunctionDef", role.Function, role.Declaration, role.Name, role.Identifier, role.Incomplete),
	AnnotateType("Lambda", MapObj(Obj{
		"body": Var("body_stmts"),
	}, Obj{
		"body": Obj{
			uast.KeyType:  String("FunctionDef.body"),
			uast.KeyRoles: funcBodyRoles,
			"body_stmts":  Var("body_stmts"),
		},
	}), role.Function, role.Declaration, role.Value, role.Anonymous),

	// Formal Arguments
	// Incomplete because we've no role for "virtual" grouping nodes
	AnnotateType("arguments", nil, role.Function, role.Declaration, role.Argument, role.Incomplete),
	AnnotateType("arg",
		FieldRoles{
			"default":    {Opt: true, Roles: role.Roles{role.Argument, role.Default}},
			"annotation": {Opt: true, Roles: role.Roles{role.Annotation, role.Noop}},
		}, role.Function, role.Declaration, role.Argument, role.Name),
	// Incomplete because we've no role for "mandatory name"
	AnnotateType("kwonly_arg",
		FieldRoles{
			"default":    {Opt: true, Roles: role.Roles{role.Argument, role.Default}},
			"annotation": {Opt: true, Roles: role.Roles{role.Annotation, role.Noop}},
		}, role.Function, role.Declaration, role.Argument, role.Name, role.Incomplete),
	AnnotateType("kwarg", nil, role.Function, role.Declaration, role.ArgsList, role.Map, role.Name),
	AnnotateType("vararg", nil, role.Function, role.Declaration, role.ArgsList, role.List, role.Name),

	// Function Calls
	AnnotateType("Call", FieldRoles{
		"args":     {Arr: true, Roles: role.Roles{role.Function, role.Call, role.Positional, role.Argument, role.Name}},
		"func":     {Roles: role.Roles{role.Call, role.Callee}},
		"keywords": {Arr: true, Roles: role.Roles{role.Function, role.Call, role.Argument}},
	}, role.Function, role.Call, role.Expression),

	// Keywords are additionally annotated in FunctionDef and ClassDef
	AnnotateType("keyword", FieldRoles{
		"value": {Roles: role.Roles{role.Argument, role.Value}},
		"arg":   {Rename: uast.KeyToken},
	}, role.Name),

	// Comments and non significative whitespace
	AnnotateType("SameLineNoops", nil, role.Comment),

	AnnotateType("PreviousNoops", FieldRoles{
		"lines": {Arr: true, Roles: role.Roles{role.Noop}},
	}, role.Noop),

	AnnotateType("RemainderNoops", FieldRoles{
		"lines": {Arr: true, Roles: role.Roles{role.Noop}},
	}, role.Noop),

	AnnotateType("NoopLine", MapObj(Obj{
		"noop_line": Var("txt"),
	}, Obj{
		uast.KeyToken: Var("txt"),
	}), role.Comment, role.Noop),

	AnnotateType("NoopSameLine", MapObj(Obj{
		"s": Var("txt"),
	}, Obj{
		uast.KeyToken: Var("txt"),
	}), role.Comment, role.Noop),

	// Import
	AnnotateType("Import", nil, role.Import, role.Declaration, role.Statement),
	AnnotateType("Import", MapObj(Obj{
		"names": Var("names"),
	}, Obj{
		"names": Obj{
			uast.KeyType: String("ImportFrom.names"),
			// Incomplete because it's a grouping node
			uast.KeyRoles: Roles(role.Import, role.Pathname, role.Identifier, role.Incomplete),
			"name_list":   Var("names"),
		},
		uast.KeyToken: String("import"),
	}), role.Import, role.Declaration, role.Statement),

	AnnotateType("ImportFrom", MapObj(Obj{
		"module": Var("module"),
		"level":  OpLevelDotsNumConv{op: Var("level"), orig: Var("origlevel"), prefix: "."},
		"names":  Var("names"),
	}, Obj{
		"names": Obj{
			uast.KeyType: String("ImportFrom.names"),
			// Incomplete because it's a grouping node
			uast.KeyRoles: Roles(role.Import, role.Pathname, role.Identifier, role.Incomplete),
			"name_list":   Var("names"),
		},
		"level": Obj{
			uast.KeyType:  String("ImportFrom.level"),
			uast.KeyToken: Var("level"),
			uast.KeyRoles: Roles(role.Import, role.Incomplete),
		},
		"module": Obj{
			uast.KeyType:  String("ImportFrom.module"),
			uast.KeyToken: Var("module"),
			uast.KeyRoles: Roles(role.Import, role.Pathname, role.Identifier),
		},
		"num_level": Var("origlevel"),
	}), role.Import, role.Declaration, role.Statement),

	AnnotateType("alias", MapObj(Obj{
		"asname": Var("asname"),
		"name":   Var("name"),
	}, Obj{
		"asname": Obj{
			uast.KeyType:  String("alias.asname"),
			uast.KeyRoles: Roles(role.Import, role.Pathname, role.Identifier, role.Alias),
			uast.KeyToken: Var("asname"),
		},
		uast.KeyToken: Var("name"),
	}), role.Import, role.Pathname, role.Identifier),

	// Class Definitions
	AnnotateType("ClassDef", MapObj(Obj{
		"decorator_list": Var("decors"),
		"body":           Var("body_stmts"),
		"bases":          Var("bases"),
		"name":           Var("name"),
	}, Obj{
		uast.KeyToken: Var("name"),
		"decorator_list": Obj{
			uast.KeyType:  String("ClassDef.decorator_list"),
			uast.KeyRoles: Roles(role.Type, role.Declaration, role.Call, role.Incomplete),
			"decorators":  Var("decors"),
		},
		"body": Obj{
			uast.KeyType:  String("ClassDef.body"),
			uast.KeyRoles: Roles(role.Type, role.Declaration, role.Body),
			"body_stmts":  Var("body_stmts"),
		},
		"bases": Obj{
			uast.KeyType:  String("ClassDef.bases"),
			uast.KeyRoles: Roles(role.Type, role.Declaration, role.Base),
			"bases":       Var("bases"),
		},
	}), role.Type, role.Declaration, role.Identifier, role.Statement),

	AnnotateType("ClassDef", FieldRoles{
		"keywords": {Arr: true, Roles: role.Roles{role.Incomplete}},
	}),

	// These two (exec & print) are AST nodes in Python2 but we convert them to functions
	// in the UAST like they are in Python3
	AnnotateType("Exec", FieldRoles{
		"body":        {Roles: role.Roles{role.Call, role.Argument, role.Positional}},
		"globals":     {Roles: role.Roles{role.Call, role.Argument, role.Positional}},
		"locals":      {Roles: role.Roles{role.Call, role.Argument, role.Positional}},
		uast.KeyToken: {Add: true, Op: String("exec")},
	}, role.Function, role.Call, role.Expression),

	AnnotateType("Print", FieldRoles{
		"values":      {Arr: true, Roles: role.Roles{role.Call, role.Argument, role.Positional}},
		uast.KeyToken: {Add: true, Op: String("print")},
	}, role.Function, role.Call, role.Callee, role.Identifier, role.Expression),

	// If and IfExpr
	AnnotateType("If", MapObj(Obj{
		"body":   Var("body_stmts"),
		"orelse": Var("else_stmts"),
		"test":   ObjectRoles("test"),
	}, Obj{
		"body": Obj{
			uast.KeyType:  String("If.body"),
			uast.KeyRoles: Roles(role.If, role.Body, role.Then),
			"body_stmts":  Var("body_stmts"),
		},
		"orelse": Obj{
			uast.KeyType:  String("If.orelse"),
			uast.KeyRoles: Roles(role.If, role.Body, role.Else),
			"else_stmts":  Var("else_stmts"),
			uast.KeyToken: String("else"),
		},
		"test": ObjectRoles("test", role.If, role.Condition),
	}), role.If, role.Expression),

	AnnotateType("IfExp", ObjRoles{
		"body":   {role.If, role.Body, role.Then},
		"test":   {role.If, role.Condition},
		"orelse": {role.If, role.Body, role.Else},
	}, role.If, role.Expression),

	// For, AsyncFor and While
	loopAnnotate("For", role.For, role.For, role.Iterator, role.Statement),
	loopAnnotate("AsyncFor", role.For, role.For, role.Iterator, role.Statement, role.Incomplete),
	loopAnnotate("While", role.While, role.While, role.Statement),
	AnnotateType("For", ObjRoles{
		"iter":   {role.For, role.Expression},
		"target": {role.For, role.Update},
	}),
	AnnotateType("AsyncFor", ObjRoles{
		"iter":   {role.For, role.Expression},
		"target": {role.For, role.Update},
	}),
	AnnotateType("While", ObjRoles{
		"test": {role.While, role.Condition},
	}),

	// Comparison nodes in Python are oddly structured. Probably one if the first
	// things that could be changed once we can normalize tree structures. Check:
	// https://greentreesnakes.readthedocs.io/en/latest/nodes.html#Compare
	AnnotateType("Compare", MapObj(Obj{
		"ops":         Var("ops"),
		"comparators": Var("comparators"),
	}, Obj{
		"ops": Obj{
			uast.KeyType:  String("Compare.ops"),
			uast.KeyRoles: Roles(role.Expression),
			"ops":         Var("ops"),
		},
		"comparators": Obj{
			uast.KeyType:  String("Compare.comparators"),
			uast.KeyRoles: Roles(role.Expression, role.Right),
			"comparators": Var("comparators"),
		},
	}), role.Expression, role.Binary, role.Condition),

	AnnotateType("Compare", ObjRoles{
		"left": {role.Expression, role.Left},
	}),
}
