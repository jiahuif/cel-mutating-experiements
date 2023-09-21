package mutator

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

var ObjectMutatorType = cel.ObjectType("kubernetes.ObjectMutator", traits.IndexerType)

var ErrNotObject = fmt.Errorf("not an object")
var ErrKeyNotFound = fmt.Errorf("key not found")

type objectMutator struct {
	object map[string]any

	abstractMutator
}

func (o *objectMutator) RemoveChild(identifier any) error {
	if s, ok := identifier.(string); ok {
		delete(o.object, s)
		return nil
	}
	return fmt.Errorf("identifier has wrong type, expect string but got %t", identifier)
}

func (o *objectMutator) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return nil, fmt.Errorf("disallowed conversion from %q to %q", ObjectMutatorType.TypeName(), typeDesc.Name())
}

func (o *objectMutator) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case ObjectMutatorType:
		return o
	case types.TypeType:
		return ObjectMutatorType
	}
	return types.NewErr("disallowed conversion from %q to %q", ObjectMutatorType.TypeName(), typeValue.TypeName())
}

func (o *objectMutator) Type() ref.Type {
	return ObjectMutatorType
}

func (o *objectMutator) Value() any {
	return types.NoSuchOverloadErr()
}

var _ Interface = (*objectMutator)(nil)
var _ Container = (*objectMutator)(nil)

func (o *objectMutator) Get(index ref.Val) ref.Val {
	f, ok := index.(types.String)
	if !ok {
		return types.MaybeNoSuchOverloadErr(f)
	}
	key := f.Value().(string)
	if v, exists := o.object[f.Value().(string)]; exists {
		switch v.(type) {
		case string:
			return types.String(v.(string))
		case int:
			return types.Int(v.(int))
		case int64:
			return types.Int(v.(int64))
		case map[string]any:
			mutator, err := NewObjectMutator(o, key)
			if err != nil {
				return types.WrapErr(err)
			}
			return mutator
		default:
			return types.NewErr("missing mutator for: %v", v)
		}
	}
	return types.NewErr("no such key: %s", f)
}

func (o *objectMutator) Merge(rhs any) ref.Val {
	patch, ok := rhs.(map[ref.Val]ref.Val)
	if !ok {
		return types.NoSuchOverloadErr()
	}
	return mergeObject(o.object, patch)
}

func (o *objectMutator) Remove() ref.Val {
	if container, ok := o.Parent().(Container); ok && o.Identifier() != nil {
		err := container.RemoveChild(o.Identifier())
		if err != nil {
			return types.WrapErr(err)
		}
		return types.NullValue
	}
	return types.NoSuchOverloadErr()
}

func NewRootObjectMutator(root map[string]any) Interface {
	mutator := new(objectMutator)
	mutator.object = root
	return mutator
}

func NewObjectMutator(parent Interface, key string) (Interface, error) {
	parentMutator, ok := parent.(*objectMutator)
	if !ok {
		return nil, ErrNotObject
	}
	field, ok := parentMutator.object[key]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrKeyNotFound, key)
	}
	object, ok := field.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNotObject, key)
	}
	mutator := new(objectMutator)
	mutator.parent = parent
	mutator.object = object
	mutator.identifier = key
	return mutator, nil
}

var _ traits.Indexer = (*objectMutator)(nil)

func mergeObject(lhs map[string]any, rhs map[ref.Val]ref.Val) ref.Val {
	for key := range rhs {
		name, ok := key.Value().(string)
		if !ok {
			return types.NewErr("bad map key: %v", key.Value())
		}
		val := rhs[key].Value()
		switch val.(type) {
		case []ref.Val:
			return types.NewErr("array cannot merge with object")
		case map[ref.Val]ref.Val:
			lhs[name] = refMapToNative(val.(map[ref.Val]ref.Val))
		default:
			lhs[name] = val
		}
	}
	return types.Null(0)
}

func refMapToNative(refMap map[ref.Val]ref.Val) map[string]any {
	ret := make(map[string]any)
	for kv, vv := range refMap {
		v := vv.Value()
		switch v.(type) {
		case []ref.Val:
			v = refSliceToNative(v.([]ref.Val))
		case map[ref.Val]ref.Val:
			v = refMapToNative(v.(map[ref.Val]ref.Val))
		}
		ret[kv.Value().(string)] = v
	}
	return ret
}

func refSliceToNative(refSlice []ref.Val) []any {
	ret := make([]any, 0, len(refSlice))
	for _, vv := range refSlice {
		v := vv.Value()
		switch v.(type) {
		case []ref.Val:
			v = refSliceToNative(v.([]ref.Val))
		case map[ref.Val]ref.Val:
			v = refMapToNative(v.(map[ref.Val]ref.Val))
		}
		ret = append(ret, v)
	}
	return ret
}
