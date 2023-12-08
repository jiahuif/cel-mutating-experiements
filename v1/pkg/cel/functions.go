package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/jiahuif/cel-mutating-experiments/v1/pkg/mutator"
)

const overloadNameObjectMerge = "mutator_object_merge"
const overloadNameObjectRemove = "mutator_object_remove"
const overloadNameListMerge = "mutator_list_merge"
const overloadNameListRemove = "mutator_list_remove"

func MergeOperation(lhs, rhs ref.Val) ref.Val {
	mutator, ok := lhs.(mutator.Interface)
	if !ok {
		return types.NoSuchOverloadErr()
	}
	return mutator.Merge(rhs.Value())
}

func RemoveOperation(lhs ref.Val) ref.Val {
	mutator, ok := lhs.(mutator.Interface)
	if !ok {
		return types.NoSuchOverloadErr()
	}
	return mutator.Remove()
}

func EnvOpts() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function("merge",
			cel.MemberOverload(overloadNameObjectMerge,
				[]*cel.Type{mutator.ObjectMutatorType, cel.AnyType},
				mutator.ObjectMutatorType, cel.BinaryBinding(MergeOperation)),
		),
		cel.Function("merge",
			cel.MemberOverload(overloadNameListMerge,
				[]*cel.Type{mutator.ListMutatorType, cel.AnyType},
				mutator.ListMutatorType, cel.BinaryBinding(MergeOperation)),
		),
		cel.Function("remove",
			cel.MemberOverload(overloadNameObjectRemove,
				[]*cel.Type{mutator.ObjectMutatorType},
				mutator.ObjectMutatorType,
				cel.UnaryBinding(RemoveOperation),
			)),
	}
}
