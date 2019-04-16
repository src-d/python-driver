package normalizer

import (
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

// FIXME: not reversible
type OpPrependPath struct {
	numLevel Op
	path     Op
	joined   Op
	prefix   string
}

func (op OpPrependPath) Kinds() nodes.Kind {
	return nodes.KindString
}

func (op OpPrependPath) Check(st *State, n nodes.Node) (bool, error) {
	v, ok := n.(nodes.Value)
	if !ok {
		return false, nil
	}

	res1, err := op.numLevel.Check(st, n)
	if err != nil || !res1 {
		return false, err
	}

	res2, err := op.path.Check(st, v)
	if err != nil || !res2 {
		return false, err
	}

	return res1 && res2, nil
}

func (op OpPrependPath) Construct(st *State, n nodes.Node) (nodes.Node, error) {
	first, err := op.numLevel.Construct(st, n)
	if err != nil || first == nil {
		return n, err
	}
	firstVal := first.(nodes.Value)
	prependVal := num2dots(firstVal, op.prefix)

	path, err := op.path.Construct(st, n)
	if err != nil || path == nil {
		return n, err
	}

	joined := string(prependVal.(nodes.String)) + string(path.(nodes.String))
	op.joined = String(joined)

	v, err := op.joined.Construct(st, n)
	if err != nil {
		return nil, err
	}

	return v, nil
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
