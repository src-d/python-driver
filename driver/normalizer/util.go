package normalizer

import (
	"github.com/bblfsh/sdk/v3/uast"
	"strings"

	"github.com/bblfsh/sdk/v3/uast/nodes"
	. "github.com/bblfsh/sdk/v3/uast/transformer"
)

func num2dots(n nodes.Value, prefix string) nodes.Value {
	if intval, ok := n.(nodes.Int); ok {
		i64val := int(intval)
		return nodes.String(strings.Repeat(prefix, int(i64val)))
	}
	return n
}

type OpSplitPath struct {
	numLevel Op
	path     Op
}

func (op OpSplitPath) Kinds() nodes.Kind {
	return nodes.KindObject
}

func (op OpSplitPath) Check(st *State, n nodes.Node) (bool, error) {
	panic("TODO") // TODO: not reversible
	return true, nil
}

func (op OpSplitPath) Construct(st *State, n nodes.Node) (nodes.Node, error) {
	var idents []uast.Identifier
	if op.numLevel != nil {
		nd, err := op.numLevel.Construct(st, nil)
		if err != nil {
			return nil, err
		}
		levels, ok := nd.(nodes.Int)
		if !ok {
			return nil, ErrUnexpectedType.New(nodes.Int(0), nd)
		}
		for i := 0; i < int(levels); i++ {
			idents = append(idents, uast.Identifier{Name: ".."})
		}
	}

	nd, err := op.path.Construct(st, n)
	if err != nil {
		return nil, err
	}
	if nd == nil {
		if len(idents) == 0 {
			idents = append(idents, uast.Identifier{Name: "."})
		}
	} else {
		path, ok := nd.(nodes.String)
		if !ok {
			return nil, ErrUnexpectedType.New(nodes.String(""), nd)
		}
		for _, name := range strings.Split(string(path), ".") {
			idents = append(idents, uast.Identifier{Name: name})
		}
	}
	if len(idents) == 1 {
		return uast.ToNode(idents[0])
	}
	return uast.ToNode(uast.QualifiedIdentifier{Names: idents})
}

type OpLevelDotsNumConv struct {
	op     Op
	orig   Op
	prefix string
}

func (op OpLevelDotsNumConv) Kinds() nodes.Kind {
	return nodes.KindString | nodes.KindInt
}

func (op OpLevelDotsNumConv) Check(st *State, n nodes.Node) (bool, error) {
	v, ok := n.(nodes.Value)
	if !ok {
		return false, nil
	}

	nv := num2dots(v, op.prefix)
	res1, err := op.op.Check(st, nv)
	if err != nil || !res1 {
		return false, err
	}

	res2, err := op.orig.Check(st, v)
	if err != nil || !res2 {
		return false, err
	}

	return res1 && res2, nil
}

func (op OpLevelDotsNumConv) Construct(st *State, n nodes.Node) (nodes.Node, error) {
	n, err := op.orig.Construct(st, n)
	if err != nil {
		return nil, err
	}

	v, ok := n.(nodes.Int)
	if !ok {
		return nil, ErrExpectedValue.New(n)
	}

	return v, nil
}

var _ Op = OpLevelDotsNumConv{}
