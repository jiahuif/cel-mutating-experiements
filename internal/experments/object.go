package experments

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

var objectMutatorTypeValue = types.NewTypeValue("io.x-k8s.ObjectMutator", traits.IndexerType)

type objectMutator struct {
	typeValue *types.TypeValue

	ref map[string]any
}

func (o *objectMutator) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return nil, fmt.Errorf("disallowed conversion from %q to %q", objectMutatorTypeValue.TypeName(), typeDesc.Name())
}

func (o *objectMutator) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case objectMutatorTypeValue:
		return o
	case types.TypeType:
		return objectMutatorTypeValue
	}
	return types.NewErr("disallowed conversion from %q to %q", objectMutatorTypeValue.TypeName(), typeValue.TypeName())
}

func (o *objectMutator) Equal(other ref.Val) ref.Val {
	return types.NoSuchOverloadErr()
}

func (o *objectMutator) Type() ref.Type {
	return objectMutatorTypeValue
}

func (o *objectMutator) Value() any {
	return types.NoSuchOverloadErr()
}

var _ ref.Val = (*objectMutator)(nil)

func (o *objectMutator) Get(index ref.Val) ref.Val {
	f, ok := index.(types.String)
	if !ok {
		return types.MaybeNoSuchOverloadErr(f)
	}
	if v, exists := o.ref[f.Value().(string)]; exists {
		switch v.(type) {
		case string:
			return nil
		default:
			return types.NewErr("missing mutator for: %v", v)
		}
	}
	return types.NewErr("no such key: %s", f)
}

func newObjectMutator(ref map[string]any) *objectMutator {
	return &objectMutator{typeValue: objectMutatorTypeValue, ref: ref}
}

var _ traits.Indexer = (*objectMutator)(nil)
