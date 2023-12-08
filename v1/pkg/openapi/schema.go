package openapi

import (
	"encoding/json"
	"io"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

func LoadSchema(reader io.Reader) (*spec.Schema, error) {
	s := new(spec.Schema)
	d := json.NewDecoder(reader)
	err := d.Decode(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
