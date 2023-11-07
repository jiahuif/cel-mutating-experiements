package apply

import (
	"fmt"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

type SchemaTracker struct {
	root *spec.Schema
	path []string
}

func (t *SchemaTracker) Schema() *spec.Schema {
	var schema spec.Schema
	for _, part := range t.path {
		schema = t.root.Properties[part]
	}
	return &schema
}

func (t *SchemaTracker) Advance(fieldName string) (*SchemaTracker, error) {
	s := t.Schema()
	if _, ok := s.Properties[fieldName]; ok {
		return &SchemaTracker{root: t.root, path: append(t.path, fieldName)}, nil
	}
	return nil, fmt.Errorf("field not found: %q", fieldName)
}
