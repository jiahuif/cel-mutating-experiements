package experments

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/version"
	apiservercel "k8s.io/apiserver/pkg/cel"
	"k8s.io/apiserver/pkg/cel/environment"
	"k8s.io/apiserver/pkg/cel/lazy"

	"github.com/jiahuif/cel-mutating-experiments/pkg/mutator"
)

func TestDeploymentSidecarContainer(t *testing.T) {
	env, err := buildTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	deployment := createDeployment()
	variables := lazy.NewMapValue(variablesType)
	rootObjectMutator := mutator.NewRootObjectMutator(deployment.Object)
	activation := &testActivation{
		variables: variables,
		object:    rootObjectMutator,
	}
	_, err = compileAndRun(env, activation, `object.spec.template.spec.merge({"replicas": 3})`)
	if err != nil {
		t.Fatal(err)
	}
	v, err := compileAndRun(env, activation, "object.spec.template.spec.replicas")
	if err != nil {
		t.Fatal(err)
	}
	if i, ok := v.Value().(int64); !ok || i != 3 {
		t.Errorf("want 3 but got %v", v.Value())
	}
}

type testActivation struct {
	variables *lazy.MapValue
	object    any
}

func compileAndRun(env *cel.Env, activation *testActivation, exp string) (ref.Val, error) {
	ast, issues := env.Compile(exp)
	if issues != nil {
		return nil, fmt.Errorf("fail to compile: %v", issues)
	}
	prog, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("cannot create program: %w", err)
	}
	v, _, err := prog.Eval(activation)
	if err != nil {
		return nil, fmt.Errorf("cannot eval program: %w", err)
	}
	return v, nil
}

func buildTestEnv() (*cel.Env, error) {
	envSet, err := environment.MustBaseEnvSet(environment.DefaultCompatibilityVersion()).Extend(
		environment.VersionedOptions{
			IntroducedVersion: version.MajorMinor(1, 28),
			EnvOptions: []cel.EnvOption{
				cel.Variable("variables", variablesType.CelType()),
				cel.Variable("object", cel.DynType),
				cel.Function("merge",
					cel.MemberOverload("mutator_object_merge",
						[]*cel.Type{mutator.ObjectMutatorType, cel.AnyType},
						mutator.ObjectMutatorType,
						cel.BinaryBinding(func(lhs ref.Val, rhs ref.Val) ref.Val {
							mutator, ok := lhs.(mutator.Interface)
							if !ok {
								return types.NoSuchOverloadErr()
							}
							return mutator.Merge(rhs.Value())
						}),
					)),
			},
			DeclTypes: []*apiservercel.DeclType{
				variablesType,
			},
		})
	if err != nil {
		return nil, err
	}
	env, err := envSet.Env(environment.StoredExpressions)
	return env, err
}

var variablesType = apiservercel.NewMapType(apiservercel.StringType, apiservercel.AnyType, 0)

func init() {
	variablesType.Fields = make(map[string]*apiservercel.DeclField)
}

func (a *testActivation) ResolveName(name string) (any, bool) {
	switch name {
	case "object":
		return a.object, true
	case "variables":
		return a.variables, true
	default:
		return nil, false
	}
}

func (a *testActivation) Parent() interpreter.Activation {
	return nil
}

func createDeployment() *unstructured.Unstructured {
	u := new(unstructured.Unstructured)
	err := json.Unmarshal([]byte(deploymentJSON), u)
	if err != nil {
		panic(err)
	}
	return u
}

// deploymentJSON is obtained from
// kubectl -n test create deploy nginx --image=nginx -o=json --dry-run=client
const deploymentJSON = `
{
    "kind": "Deployment",
    "apiVersion": "apps/v1",
    "metadata": {
        "name": "nginx",
        "namespace": "test",
        "creationTimestamp": null,
        "labels": {
            "app": "nginx"
        }
    },
    "spec": {
        "replicas": 1,
        "selector": {
            "matchLabels": {
                "app": "nginx"
            }
        },
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "app": "nginx"
                }
            },
            "spec": {
                "containers": [
                    {
                        "name": "nginx",
                        "image": "nginx",
                        "resources": {}
                    }
                ]
            }
        },
        "strategy": {}
    },
    "status": {}
}
`
