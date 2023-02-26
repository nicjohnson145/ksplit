package internal

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"path/filepath"

	"github.com/fluxcd/pkg/ssa"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var combinedKinds = map[string]string{
	"Role":               "rbac.yaml",
	"ClusterRole":        "rbac.yaml",
	"RoleBinding":        "rbac.yaml",
	"ClusterRoleBinding": "rbac.yaml",
	"ServiceAccount":     "rbac.yaml",

	"CustomResourceDefinition": "crds.yaml",

	"MutatingWebhookConfiguration":   "webhook.yaml",
	"ValidatingWebhookConfiguration": "webhook.yaml",
}

func NewSplitter(file []byte) *Splitter {
	return &Splitter{
		file: file,
	}
}

type Splitter struct {
	file []byte
	pathPrefix string
}

func (s *Splitter) Split() error {
	objects, err := s.toObjects()
	if err != nil {
		return err
	}

	files := map[string][]*unstructured.Unstructured{}
	for _, obj := range objects {
		file := strings.ToLower(obj.GetKind()) + ".yaml"
		if val, ok := combinedKinds[obj.GetKind()]; ok {
			file = val
		}

		files[file] = append(files[file], obj)
	}

	for file, objects := range files {
		if err := s.writeFile(file, objects); err != nil {
			return fmt.Errorf("error writing %v: %w", file, err)
		}
	}

	return nil
}

func (s *Splitter) toObjects() ([]*unstructured.Unstructured, error) {
	objects, err := ssa.ReadObjects(bytes.NewReader(s.file))
	if err != nil {
		return nil, fmt.Errorf("error reading objects: %w", err)
	}

	return objects, nil
}

func (s *Splitter) writeFile(file string, objects []*unstructured.Unstructured) error {
	path := filepath.Join(s.pathPrefix, file)
	fl, openErr := os.Create(path)
	if openErr != nil {
		return openErr
	}
	defer fl.Close()

	var jsonBytes []byte
	var err error

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if _, err := fl.Write([]byte("---\n")); err != nil {
		return fmt.Errorf("error during write: %w", err)
	}
	for _, obj := range objects {
		jsonBytes, err = obj.MarshalJSON()
		if err != nil {
			return fmt.Errorf("error marshalling %v: %w", obj.GetName(), err)
		}

		var tmp map[string]any
		err = yaml.Unmarshal(jsonBytes, &tmp)
		if err != nil {
			return fmt.Errorf("error parsing %v as yaml: %w", obj.GetName(), err)
		}

		err = encoder.Encode(tmp)
		if err != nil {
			return fmt.Errorf("error converting to %v to yaml: %w", obj.GetName(), err)
		}

		if _, err := fl.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("error during write: %w", err)
		}
		buf.Reset()
	}

	return nil
}
