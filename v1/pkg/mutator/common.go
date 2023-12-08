package mutator

import (
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func mutatorOf(v any, parent Container, key any) ref.Val {
	switch v.(type) {
	case string:
		return types.String(v.(string))
	case int:
		return types.Int(v.(int))
	case int64:
		return types.Int(v.(int64))
	case map[string]any:
		mutator, err := NewObjectMutator(parent, key)
		if err != nil {
			return types.WrapErr(err)
		}
		return mutator
	case []any:
		mutator, err := NewListMutator(parent, key)
		if err != nil {
			return types.WrapErr(err)
		}
		return mutator
	default:
		return types.NewErr("missing mutator for: %v", v)
	}
}
