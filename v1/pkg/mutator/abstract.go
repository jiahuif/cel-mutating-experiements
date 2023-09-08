package mutator

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type abstractMutator struct {
	parent     Interface
	identifier any
}

var abstractMutatorTypeValue = cel.ObjectType("io.x-k8s.AbstractMutator")

func (a *abstractMutator) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return nil, fmt.Errorf("disallowed conversion from %q to %q", a.Type().TypeName(), typeDesc.Name())
}

func (a *abstractMutator) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NoSuchOverloadErr()
}

func (a *abstractMutator) Equal(other ref.Val) ref.Val {
	return types.NoSuchOverloadErr()
}

func (a *abstractMutator) Type() ref.Type {
	return abstractMutatorTypeValue
}

func (a *abstractMutator) Value() any {
	return types.NoSuchOverloadErr()
}

func (a *abstractMutator) Parent() Interface {
	return a.parent
}

func (a *abstractMutator) Identifier() any {
	return a.identifier
}

func (a *abstractMutator) Merge(patch any) ref.Val {
	return types.NoSuchOverloadErr()
}

func (a *abstractMutator) Remove() ref.Val {
	return types.NoSuchOverloadErr()
}

var _ Interface = (*abstractMutator)(nil)
