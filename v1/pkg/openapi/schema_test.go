package openapi

import (
	"os"
	"testing"
)

func TestLoadSchema(t *testing.T) {
	var f *os.File
	var err error
	f, err = os.Open("../../testdata/deploy.schema.json")
	if err != nil {
		f, err = os.Open("testdata/deploy.schema.json")
	}
	if f == nil {
		t.Fatalf("cannot load schema file: %v", err)
	}
	defer f.Close()
	s, err := LoadSchema(f)
	if err != nil {
		t.Fatalf("cannot load schema: %v", s)
	}
	if len(s.Type) != 1 || s.Type[0] != "object" {
		t.Errorf("unexpected type: %v", s.Type)
	}
}
