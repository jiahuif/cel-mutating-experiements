package experments

import (
	"github.com/google/cel-go/common/types"
)

var stringMutatorTypeValue = types.NewTypeValue("io.x-k8s.StringMutator")

type stringMutator struct {
	parent map[string]any
	key    string
}
