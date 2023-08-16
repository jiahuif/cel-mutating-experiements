package experments

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/cel-go/interpreter"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/version"
	apiservercel "k8s.io/apiserver/pkg/cel"
	"k8s.io/apiserver/pkg/cel/environment"
	"k8s.io/apiserver/pkg/cel/lazy"
)

func TestDeploymentSidecarContainer(t *testing.T) {
	/*
		variables:
		- name: existingContainer
			expression: object.spec.template.containers.filter(c, c.name.value() == 'sidecar')
		- name: targetContainer
			expression: object.spec.template.containers.append()
		- name: targetEnvs
			expression: variables.targetContainer.env.clear().resize(2)
		- name: simpleEnv
			expression: variables.targetEnvs[0]
		- name: envFromConfig
			expression: variables.targetEnvs[1]
		- name: configMapKeyRef
			expression: variables.envFromConfig.configMapKeyRef
		- name: resourceLimits
			expression: variables.targetContainer.resource.limits.clear()
		- name: resourceRequests
			expression: variables.targetContainer.resource.requests.clear()
		mutation:
		- condition: existingContainer.size() == 0
			expressions:
			- variables.targetContainer.image.set('example.com/sidecar:v1')
			- variables.simpleEnv.name.set('FOO')
			- variables.simpleEnv.value.set('BAR')
			- variables.envFromConfig.name.set('FROM_CONFIG')
			- variables.configMapKeyRef.name.set('demo-config')
			- variables.configMapKeyRef.key.set('demo-key')
			- variables.resourceLimits["memory"].set("128Mi")
			- variables.resourceLimits["cpu"].set("250m")
			- variables.resourceRequests["memory"].set("128Mi")
			- variables.resourceRequests["cpu"].set("250m")
	*/
	env, err := buildTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	deployment := createDeployment()
	variables := lazy.NewMapValue(variablesType)
	activation := &testActivation{
		variables: variables,
		object:    deployment.Object,
	}
	v, err := compileAndRun(env, activation, "variables")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
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
var objectMutatorTypeValue = types.NewTypeValue("io.x-k8s.ObjectMutator", traits.IndexerType)
var stringMutatorTypeValue = types.NewTypeValue("io.x-k8s.StringMutator")

func init() {
	variablesType.Fields = make(map[string]*apiservercel.DeclField)
}

type objectMutator struct {
	typeValue *types.TypeValue

	ref map[string]any
}

func newObjectMutator(ref map[string]any) *objectMutator {
	return &objectMutator{typeValue: objectMutatorTypeValue, ref: ref}
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
