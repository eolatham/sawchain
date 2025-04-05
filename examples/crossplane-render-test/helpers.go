package test

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// runCrossplaneRender runs `crossplane render` with given XR, composition, functions,
// and any number of --extra-resources files.
func runCrossplaneRender(xrPath, compositionPath, functionsPath string, extraResources ...string) (string, error) {
	args := []string{
		"render",
		xrPath,
		compositionPath,
		functionsPath,
	}

	for _, res := range extraResources {
		args = append(args, "--extra-resources="+res)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("crossplane", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run crossplane render: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// unstructuredObjectsFromYaml parses a YAML string into a slice of unstructured objects.
func unstructuredObjectsFromYaml(yamlStr string) ([]client.Object, error) {
	var result []client.Object

	// Split multi-document YAML by "---" separator
	documents := strings.Split(yamlStr, "---")

	for _, doc := range documents {
		// Skip empty documents
		if strings.TrimSpace(doc) == "" {
			continue
		}

		// Parse YAML into unstructured object
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(doc), obj); err != nil {
			return nil, err
		}

		// Skip documents without apiVersion and kind
		if obj.GetAPIVersion() == "" || obj.GetKind() == "" {
			continue
		}

		result = append(result, obj)
	}

	return result, nil
}
