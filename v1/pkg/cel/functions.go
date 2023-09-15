package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/jiahuif/cel-mutating-experiments/v1/pkg/mutator"
)

const overloadNameMerge = "mutator_object_merge"
const overloadNameRemove = "mutator_object_remove"

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
			cel.MemberOverload(overloadNameMerge,
				[]*cel.Type{mutator.ObjectMutatorType, cel.AnyType},
				mutator.ObjectMutatorType, cel.BinaryBinding(MergeOperation)),
		),
		cel.Function("remove",
			cel.MemberOverload(overloadNameRemove,
				[]*cel.Type{mutator.ObjectMutatorType},
				mutator.ObjectMutatorType,
				cel.UnaryBinding(RemoveOperation),
			)),
	}
}
