package apply

import (
	"fmt"
	"os"
	"testing"

	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/structured-merge-diff/v4/typed"

	"github.com/jiahuif/cel-mutating-experiments/v1/pkg/openapi"
)

func TestSimpleApply(t *testing.T) {
	ty, err := loadType()
	if err != nil {
		t.Fatal(err)
	}
	original, err := ty.FromStructured(&v1.Deployment{Spec: v1.DeploymentSpec{
		Replicas: ptr.To[int32](1),
	}})
	if err != nil {
		t.Fatal(err)
	}
	patch, err := ty.FromStructured(&v1.Deployment{Spec: v1.DeploymentSpec{
		Replicas: ptr.To[int32](3),
	}})
	if err != nil {
		t.Fatal(err)
	}
	mergeResult, err := original.Merge(patch)
	result := mergeResult.AsValue().Unstructured()
	if result.(map[string]any)["spec"].(map[string]any)["replicas"].(int64) != 3 {
		t.Errorf("unexpected resulting replica count")
	}
}

func TestStructuralApply(t *testing.T) {
	ty, err := loadType()
	if err != nil {
		t.Fatal(err)
	}
	original, err := ty.FromStructured(&v1.Deployment{Spec: v1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "nginx",
						Image: "nginx",
					},
					{
						Name:  "sidecar",
						Image: "example.com/sidecar:v1",
					},
				},
			},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	patch, err := ty.FromStructured(&v1.Deployment{Spec: v1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "sidecar",
						Image: "example.com/sidecar:v2",
					},
				},
			},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	mergeResult, err := original.Merge(patch)
	result := mergeResult.AsValue().Unstructured()
	sidecarContainer := result.(map[string]any)["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)[1].(map[string]any)
	if sidecarContainer["image"].(string) != "example.com/sidecar:v2" {
		t.Errorf("unexpected image value: %s", sidecarContainer["image"].(string))
	}
}

func loadType() (*typed.ParseableType, error) {
	var f *os.File
	var err error
	f, err = os.Open("../../testdata/deploy.schema.json")
	if err != nil {
		f, err = os.Open("testdata/deploy.schema.json")
	}
	if f == nil {
		return nil, fmt.Errorf("cannot load schema file: %w", err)
	}
	defer f.Close()
	s, err := openapi.LoadSchema(f)
	if err != nil {
		return nil, fmt.Errorf("cannot load schema: %w", err)
	}
	return CreateObjectType(s)
}
