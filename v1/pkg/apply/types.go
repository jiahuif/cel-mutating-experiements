package apply

import (
	"k8s.io/kube-openapi/pkg/schemaconv"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/structured-merge-diff/v4/schema"
	"sigs.k8s.io/structured-merge-diff/v4/typed"
)

func CreateObjectType(objectSchema *spec.Schema) (*typed.ParseableType, error) {
	schemas, err := schemaconv.ToSchemaFromOpenAPI(map[string]*spec.Schema{
		"object": objectSchema,
	}, false)
	if err != nil {
		return nil, err
	}
	parser := typed.Parser{Schema: schema.Schema{Types: schemas.Types}}
	t := parser.Type("object")
	return &t, nil
}
