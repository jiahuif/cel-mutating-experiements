package integration

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/apimachinery/pkg/util/version"
	apiservercel "k8s.io/apiserver/pkg/cel"
	"k8s.io/apiserver/pkg/cel/environment"
	"k8s.io/apiserver/pkg/cel/lazy"

	"github.com/jiahuif/cel-mutating-experiments/v1/pkg/api"
	mutatorcel "github.com/jiahuif/cel-mutating-experiments/v1/pkg/cel"
	"github.com/jiahuif/cel-mutating-experiments/v1/pkg/mutator"
)

func unmarshallTestData(t *testing.T, fileName string, v any) {
	f, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("fail to load test data: %v", err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("fail to read test data: %v", err)
	}
	err = yaml.Unmarshal(b, v)
	if err != nil {
		t.Fatalf("fail to parse test data: %v", err)
	}
}

func guessTestDataFileNames(baseName string) (deploy, mutation, expected string, err error) {
	prefix := "testdata"
	if _, err := os.Stat(fmt.Sprintf("%s/%s/deploy.yaml", prefix, baseName)); err == nil {
		return fmt.Sprintf("%s/%s/deploy.yaml", prefix, baseName),
			fmt.Sprintf("%s/%s/mutation.yaml", prefix, baseName),
			fmt.Sprintf("%s/%s/expected.yaml", prefix, baseName), nil
	}
	prefix = "../../testdata"
	if _, err := os.Stat(fmt.Sprintf("%s/%s/deploy.yaml", prefix, baseName)); err == nil {
		return fmt.Sprintf("%s/%s/deploy.yaml", prefix, baseName),
			fmt.Sprintf("%s/%s/mutation.yaml", prefix, baseName),
			fmt.Sprintf("%s/%s/expected.yaml", prefix, baseName), nil
	}
	return "", "", "", fmt.Errorf("unable to locate test data for %q", baseName)
}

func runTestFromFile(t *testing.T, baseName string) {
	deployFileName, mutationFileName, expectedFileName, err := guessTestDataFileNames(baseName)
	if err != nil {
		t.Fatalf("missing input for test case %q", baseName)
	}
	variables := lazy.NewMapValue(variablesType)
	deploy := new(unstructured.Unstructured)
	unmarshallTestData(t, deployFileName, deploy)
	rootObjectMutator := mutator.NewRootObjectMutator(deploy.Object)
	mutation := new(api.MutatingAdmissionPolicy)
	unmarshallTestData(t, mutationFileName, mutation)
	expectedDeploy := new(unstructured.Unstructured)
	unmarshallTestData(t, expectedFileName, expectedDeploy)
	activation := &testActivation{
		variables: variables,
		object:    rootObjectMutator,
	}
	env, err := buildTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mutation.Spec.Mutation {
		for _, e := range m.Expressions {
			_, err := compileAndRun(env, activation, e)
			if err != nil {
				t.Fatalf("fail to eval: %v", err)
			}
		}
	}
	if !reflect.DeepEqual(deploy, expectedDeploy) {
		t.Errorf("wrong result, expected\n%v\n but got \n%v\n", expectedDeploy, deploy.Object)
	}
}

func TestSimpleMerge(t *testing.T) {
	runTestFromFile(t, "simplemerge")
}

func TestSimpleRemove(t *testing.T) {
	runTestFromFile(t, "simpleremove")
}

func TestObjectMerge(t *testing.T) {
	runTestFromFile(t, "objectmerge")
}

func TestListMerge(t *testing.T) {
	runTestFromFile(t, "listmerge")
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
			EnvOptions: append([]cel.EnvOption{
				cel.Variable("variables", variablesType.CelType()),
				cel.Variable("object", cel.DynType),
			}, mutatorcel.EnvOpts()...),
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
