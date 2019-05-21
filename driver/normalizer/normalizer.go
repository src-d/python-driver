package normalizer

import (
	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/role"
	. "github.com/bblfsh/sdk/v3/uast/transformer"
	"github.com/bblfsh/sdk/v3/uast/transformer/positioner"
)

var Preprocess = Transformers([][]Transformer{
	{Mappings(Preprocessors...)},
}...)

var PreprocessCode = []CodeTransformer{
	positioner.FromLineCol(),
}

var Preprocessors = []Mapping{
	ObjectToNode{
		InternalTypeKey: "ast_type",
		LineKey:         "lineno",
		ColumnKey:       "col_offset",
		EndLineKey:      "end_lineno",
		EndColumnKey:    "end_col_offset",
	}.Mapping(),
}

var Normalize = Transformers([][]Transformer{
	{Mappings(Normalizers...)},
}...)

func funcDefMap(typ string, async bool) Mapping {
	return MapSemantic(typ, uast.FunctionGroup{}, MapObj(
		Fields{
			{Name: "body", Op: Var("body")},
			{Name: "name", Op: Var("name")},
			// Arguments should be converted by the uast.Arguments normalization
			{Name: "args", Op: Obj{
				"args":       Var("arguments"),
				uast.KeyPos:  Var("_pos"),
				uast.KeyType: Var("_type"),
			}},
			// Will be filled only if there is a Python3 type annotation. If the field is
			// missing or the value is nil it DOESNT mean that the function returns None,
			// just that there is no annotation (but the function could still return some
			// other type!).
			{Name: "returns", Optional: "ret_opt", Op: Cases("ret_case",
				// case 1: no type annotations for the return
				Is(nil),
				// case 2: boxed identifier
				Fields{
					{Name: uast.KeyType, Op: String("BoxedName")},
					{Name: "boxed_value", Op: Var("ret_type")},
					// No problem dropping this one, it's used by an internal interpreter optimization/cache
					// without semantic meaning
					{Name: "ctx", Op: Any()},
					// FIXME: change this once we've a way to store other nodes on semantic objects
					// See: https://github.com/bblfsh/sdk/issues/361
					// See: https://github.com/bblfsh/python-driver/issues/178
					{Name: "noops_previous", Drop: true, Op: Any()},
					{Name: "noops_sameline", Drop: true, Op: Any()},
				},
				// case 3: everything else
				// TODO: not reversible
				Var("ret_type"),
			)},
			{Name: "decorator_list", Op: Var("func_decorators")},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		},
		Obj{
			"Nodes": Arr(
				Fields{
					// FIXME: generator=true if it uses yield anywhere in the body
					{Name: "async", Op: Bool(async)},
					{Name: "decorators", Op: Var("func_decorators")},
					{Name: "comments", Op: Fields{
						{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
						{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
					}},
				},
				UASTType(uast.Alias{}, Obj{
					// FIXME: can't call identifierWithPos because it would take the position of the
					//        function node that is not exactly the same as the position of the function name
					"Name": UASTType(uast.Identifier{}, Obj{
						"Name": Var("name"),
					}),
					"Node": UASTType(uast.Function{}, Obj{
						"Type": UASTType(uast.FunctionType{}, Fields{
							{Name: "Arguments", Op: Var("arguments")},
							{Name: "Returns", Optional: "ret_opt", Op: Cases("ret_case",
								// Python always adds an implicit return of None if function
								// returns nothing, so we will set Init on our returns to None.

								// case 1: no type annotations for the return
								Arr(UASTType(uast.Argument{},
									Obj{
										"Init": UASTType(uast.Identifier{},
											Obj{"Name": String("None")},
										),
									},
								)),
								// case 2: boxed identifier
								Arr(UASTType(uast.Argument{},
									Obj{
										"Type": Var("ret_type"),
										"Init": UASTType(uast.Identifier{},
											Obj{"Name": String("None")},
										),
									},
								)),
								// case 3: everything else
								Arr(UASTType(uast.Argument{},
									Obj{
										"Type": Var("ret_type"),
										"Init": UASTType(uast.Identifier{},
											Obj{"Name": String("None")},
										),
									},
								)),
							)},
						}),
						"Body": UASTType(uast.Block{}, Obj{
							"Statements": Var("body"),
						}),
					}),
				}),
			),
		},
	))
}

func identifierWithPos(nameVar string) ObjectOp {
	return UASTType(uast.Identifier{}, Obj{
		uast.KeyPos: UASTType(uast.Positions{}, Obj{
			uast.KeyStart: Var(uast.KeyStart),
			uast.KeyEnd:   Var(uast.KeyEnd),
		}),
		"Name": Var(nameVar),
	})
}

// mapStr factorizes the common annotation for string types (Byte, Str, StrLiteral)
func mapStr(nativeType string) Mapping {
	return Map(
		Part("_", Fields{
			{Name: uast.KeyType, Op: String(nativeType)},
			{Name: uast.KeyPos, Op: Var("pos_")},
			{Name: "s", Op: Var("s")},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
		Part("_", Fields{
			{Name: uast.KeyType, Op: String("Boxed" + nativeType)},
			{Name: "boxed_value", Op: UASTType(uast.String{}, Obj{
				uast.KeyPos: Var("pos_"),
				"Value":     Var("s"),
			})},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
	)
}

var Normalizers = []Mapping{

	// Box Names, Strings, Attributes, and Bools into a "BoxedFoo" moving the real node to the
	// "value" property and keeping the comments in the parent (if not, comments would be lost
	// when promoting the objects).
	// For other objects, the comments are dropped.
	// See: https://github.com/bblfsh/sdk/issues/361
	Map(
		Part("_", Fields{
			{Name: uast.KeyType, Op: String("Name")},
			{Name: uast.KeyPos, Op: Var("pos_")},
			{Name: "id", Op: Var("id")},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
		Part("_", Fields{
			{Name: uast.KeyType, Op: String("BoxedName")},
			{Name: "boxed_value", Op: UASTType(uast.Identifier{}, Obj{
				uast.KeyPos: Var("pos_"),
				"Name":      Var("id"),
			})},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
	),

	Map(
		Part("_", Fields{
			{Name: uast.KeyType, Op: String("BoolLiteral")},
			{Name: uast.KeyPos, Op: Var("pos_")},
			{Name: "value", Op: Var("lv")},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
		Part("_", Fields{
			{Name: uast.KeyType, Op: String("BoxedBoolLiteral")},
			{Name: "boxed_value", Op: UASTType(uast.Bool{}, Obj{
				uast.KeyPos: Var("pos_"),
				"Value":     Var("lv"),
			})},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
	),

	Map(
		Part("_", Fields{
			{Name: uast.KeyType, Op: String("Attribute")},
			{Name: uast.KeyPos, Op: Var("pos_")},
			{Name: "attr", Op: Var("aname")},
			// No problem dropping this one, it's used by an internal interpreter optimization
			// cache without semantic meaning
			// without semantic meaning
			{Name: "ctx", Op: Any()},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
		Part("_", Fields{
			{Name: uast.KeyType, Op: String("BoxedAttribute")},
			{Name: "boxed_value", Op: UASTType(uast.Identifier{}, Obj{
				uast.KeyPos: Var("pos_"),
				"Name":      Var("aname"),
			})},
			{Name: "noops_previous", Optional: "np_opt", Op: Var("noops_previous")},
			{Name: "noops_sameline", Optional: "ns_opt", Op: Var("noops_sameline")},
		}),
	),

	mapStr("Bytes"),
	mapStr("Str"),
	mapStr("StringLiteral"),

	MapSemantic("NoopLine", uast.Comment{}, MapObj(
		Obj{
			"noop_line": CommentTextTrimmed([2]string{"#", ""}, "comm"),
		},
		CommentNode(false, "comm", nil),
	)),

	MapSemantic("NoopSameLine", uast.Comment{}, MapObj(
		Obj{
			"s": CommentText([2]string{"#", ""}, "comm"),
		},
		CommentNode(false, "comm", nil),
	)),

	// remove empty noops
	Map(Obj{
		uast.KeyType: String("NoopSameLine"),
		uast.KeyPos:  Any(),
		"s":          String(""),
	}, Is(nil)),
	Map(Obj{
		uast.KeyType: String("SameLineNoops"),
		uast.KeyPos:  Any(),
		"noop_lines": Check(All(Is(nil)), Any()),
	}, Is(nil)),

	// FIXME: no positions for keywords in the native AST
	AnnotateType("keyword", MapObj(
		Fields{
			{Name: "arg", Op: Var("name")},
			// FIXME: change this once we've a way to store other nodes on semantic objects
			// See: https://github.com/bblfsh/sdk/issues/361
			// See: https://github.com/bblfsh/python-driver/issues/178
			{Name: "noops_previous", Drop: true, Op: Any()},
			{Name: "noops_sameline", Drop: true, Op: Any()},
		},
		Fields{
			{Name: "arg",
				Op: UASTType(uast.Identifier{}, Obj{
					"Name": Var("name"),
				})},
		}),
		role.Name),

	MapSemantic("arg", uast.Argument{}, MapObj(
		Fields{
			{Name: uast.KeyToken, Op: Var("name")},
			{Name: "default", Optional: "opt_def", Op: Var("init")},
			// No problem dropping this one, it's used by an internal interpreter optimization/cache
			// without semantic meaning
			{Name: "ctx", Optional: "opt_ctx", Op: Any()},
			// FIXME: change this once we've a way to store other nodes on semantic objects
			// See: https://github.com/bblfsh/sdk/issues/361
			// See: https://github.com/bblfsh/python-driver/issues/178
			{Name: "noops_previous", Drop: true, Op: Any()},
			{Name: "noops_sameline", Drop: true, Op: Any()},
			// This one is pesky - they're ignored by the runtime, could have typing from
			// mypy, or could have anything else, so we can assign to the semantic type
			{Name: "annotation", Optional: "ann_opt", Op: Any()},
		},
		Fields{
			{Name: "Name", Op: identifierWithPos("name")},
			{Name: "Init", Optional: "opt_def", Op: Var("init")},
		},
	)),

	MapSemantic("kwonly_arg", uast.Argument{}, MapObj(
		Fields{
			{Name: uast.KeyToken, Op: Var("name")},
			{Name: "default", Op: Var("init")},
			// FIXME: change this once we've a way to store other nodes on semantic objects
			// See: https://github.com/bblfsh/sdk/issues/361
			// See: https://github.com/bblfsh/python-driver/issues/178
			{Name: "noops_previous", Drop: true, Op: Any()},
			{Name: "noops_sameline", Drop: true, Op: Any()},
			// This one is pesky - they're ignored by the runtime, could have typing from
			// mypy, or could have anything else, so we can assign to the semantic type
			{Name: "annotation", Op: Any()},
		},
		Obj{
			"Init": Var("init"),
			"Name": identifierWithPos("name"),
		},
	)),

	MapSemantic("vararg", uast.Argument{}, MapObj(
		Fields{
			{Name: uast.KeyToken, Op: Var("name")},
			// FIXME: change this once we've a way to store other nodes on semantic objects
			// See: https://github.com/bblfsh/sdk/issues/361
			// See: https://github.com/bblfsh/python-driver/issues/178
			{Name: "noops_previous", Drop: true, Op: Any()},
			{Name: "noops_sameline", Drop: true, Op: Any()},
			// This one is pesky - they're ignored by the runtime, could have typing from
			// mypy, or could have anything else, so we can assign to the semantic type
			{Name: "annotation", Op: Any()},
		},
		Obj{
			"Name":     identifierWithPos("name"),
			"Variadic": Bool(true),
		},
	)),

	MapSemantic("kwarg", uast.Argument{}, MapObj(
		Fields{
			{Name: uast.KeyToken, Op: Var("name")},
			// FIXME: change this once we've a way to store other nodes on semantic objects
			// See: https://github.com/bblfsh/sdk/issues/361
			// See: https://github.com/bblfsh/python-driver/issues/178
			{Name: "noops_previous", Drop: true, Op: Any()},
			{Name: "noops_sameline", Drop: true, Op: Any()},
			// This one is pesky - they're ignored by the runtime, could have typing from
			// mypy, or could have anything else, so we can assign to the semantic type
			{Name: "annotation", Op: Any()},
		},
		Obj{
			"Name":        identifierWithPos("name"),
			"MapVariadic": Bool(true),
		},
	)),

	funcDefMap("FunctionDef", false),
	funcDefMap("AsyncFunctionDef", true),

	AnnotateType("Import", MapObj(
		Obj{
			"names": Each("vals", Var("name")),
		},
		Obj{
			"names": Each("vals", UASTType(uast.RuntimeImport{},
				Obj{"Path": Var("name")},
			)),
		},
	), role.Import, role.Declaration, role.Statement),

	// pre-process aliases: remove ones with no alias (useless)
	Map(Obj{
		uast.KeyType: String("alias"),
		uast.KeyPos:  Any(),
		"name":       Check(OfKind(nodes.KindString), Var("name")),
		"asname":     Is(nil),
	}, UASTType(uast.Identifier{}, Obj{
		uast.KeyPos: UASTType(uast.Positions{}, nil),
		"Name":      Var("name"),
	})),

	// FIXME: aliases doesn't have a position (can't be currently fixed by the tokenizer
	//        because they don't even have a line in the native AST)
	MapSemantic("alias", uast.Alias{}, MapObj(
		Obj{
			"name":   Var("name"),
			"asname": Var("alias"),
		},
		Obj{
			"Name": UASTType(uast.Identifier{}, Obj{
				"Name": Var("name"),
			}),
			"Node": UASTType(uast.Identifier{},
				Obj{"Name": Var("alias")},
			),
		},
	)),

	// Star imports
	MapSemantic("ImportFrom", uast.RuntimeImport{}, MapObj(
		Fields{
			{Name: "names", Op: Arr(
				Obj{
					uast.KeyType: String("uast:Identifier"),
					uast.KeyPos:  Any(),
					"Name":       String("*"),
				},
			)},
			{Name: "level", Op: Var("level")},
			{Name: "module", Op: Var("module")},
			// FIXME: change this once we've a way to store other nodes on semantic objects
			// See: https://github.com/bblfsh/sdk/issues/361
			// See: https://github.com/bblfsh/python-driver/issues/178
			{Name: "noops_previous", Drop: true, Op: Any()},
			{Name: "noops_sameline", Drop: true, Op: Any()},
		},
		Obj{
			"All": Bool(true),
			"Path": UASTType(uast.Identifier{}, Obj{
				"Name": OpPrependPath{
					// FIXME: no position for the module (path) in the native AST, only when the import starts
					numLevel: Var("level"),
					path:     Var("module"),
					joined:   Var("joined"),
					prefix:   "../",
				},
			}),
		},
	)),

	MapSemantic("ImportFrom", uast.RuntimeImport{}, MapObj(
		Fields{
			{Name: "names", Op: Var("names")},
			{Name: "module", Op: Var("module")},
			{Name: "level", Op: Var("level")},
			// FIXME: change this once we've a way to store other nodes on semantic objects
			// See: https://github.com/bblfsh/sdk/issues/361
			// See: https://github.com/bblfsh/python-driver/issues/178
			{Name: "noops_previous", Drop: true, Op: Any()},
			{Name: "noops_sameline", Drop: true, Op: Any()},
		},
		Obj{
			"Names": Var("names"),
			"Path": UASTType(uast.Identifier{}, Obj{
				"Name": OpPrependPath{
					numLevel: Var("level"),
					path:     Var("module"),
					joined:   Var("joined"),
					prefix:   "../",
				},
			}),
		},
	)),
}
