package mutator

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

var ListMutatorType = cel.ObjectType("kubernetes.ListMutator", traits.IndexerType)

var ErrNotList = fmt.Errorf("not a list")
var ErrListIndexOutOfBound = fmt.Errorf("index out of bound")

type listMutator struct {
	list []any

	abstractMutator
}

func NewListMutator(parent Container, key any) (Interface, error) {
	child, ok := parent.Child(key)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrKeyNotFound, key)
	}
	list, ok := child.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNotObject, key)
	}
	mutator := new(listMutator)
	mutator.parent = parent
	mutator.list = list
	mutator.identifier = key
	return mutator, nil
}

func (l *listMutator) RemoveChild(identifier any) error {
	if i, ok := identifier.(int); ok {
		if i > len(l.list) {
			return ErrListIndexOutOfBound
		}
		l.list = append(l.list[0:i], l.list[i+1:len(l.list)]...)
		return nil
	}
	return fmt.Errorf("expect index to be an int, but got a %t", identifier)
}

func (l *listMutator) Child(identifier any) (any, bool) {
	if i, ok := identifier.(int); ok {
		if i > len(l.list) {
			return nil, false
		}
		return l.list[i], true
	}
	return nil, false
}

func (l *listMutator) Get(index ref.Val) ref.Val {
	iv, ok := index.(types.Int)
	if !ok {
		return types.MaybeNoSuchOverloadErr(iv)
	}
	i := iv.Value().(int)
	if i < len(l.list) {
		v := l.list[i]
		switch v.(type) {
		case map[string]any:
			return mutatorOf(v, l, i)
		case []any:
			return mutatorOf(v, l, i)
		default:
			return types.NewErr("missing mutator for %t", v)
		}
	}
	return types.NewErr("array index out of bound: %d", i)
}

func (l *listMutator) Merge(rhs any) ref.Val {
	patch, ok := rhs.([]ref.Val)
	if !ok {
		return types.NoSuchOverloadErr()
	}
	return l.mergeList(patch)
}

func (l *listMutator) Type() ref.Type {
	return ListMutatorType
}

func (l *listMutator) SetChild(identifier any, value any) error {
	if i, ok := identifier.(int); ok {
		if i > len(l.list) {
			return ErrListIndexOutOfBound
		}
		l.list[i] = value
		return nil
	}
	return fmt.Errorf("expect index to be an int, but got a %t", identifier)
}
func (l *listMutator) mergeList(rhs []ref.Val) ref.Val {
	for _, vv := range rhs {
		var v any
		switch vv.Value().(type) {
		case []ref.Val:
			v = refSliceToNative(vv.Value().([]ref.Val))
		case map[ref.Val]ref.Val:
			v = refMapToNative(vv.Value().(map[ref.Val]ref.Val))
		default:
			v = vv.Value()
		}
		l.list = append(l.list, v)
	}
	err := l.Parent().(Container).SetChild(l.Identifier(), l.list)
	if err != nil {
		return types.WrapErr(err)
	}
	return types.NullValue
}
