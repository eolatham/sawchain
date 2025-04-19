package example

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eolatham/sawchain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	definitionsDir = "cue"

	applicationTemplatePath    = "yaml/application.tpl.yaml"
	expectedOutputTemplatePath = "yaml/expected-output.tpl.yaml"
)

var _ = Describe("Vela Dry-Run", func() {
	DescribeTable("rendering resources with KubeVela",
		func(port int, annotations map[string]string, expectedErrs []string) {
			// Create Sawchain with fake client and global bindings
			sc := sawchain.New(GinkgoTB(), fake.NewClientBuilder().Build(),
				map[string]any{"port": port, "annotations": annotations})

			// Render input template file
			applicationPath := filepath.Join(GinkgoT().TempDir(), "application.yaml")
			sc.RenderToFile(applicationPath, applicationTemplatePath)

			// Render expected output template
			expectedOutput := sc.RenderToString(expectedOutputTemplatePath)

			// Run vela dry-run
			output, err := runVelaDryRun(applicationPath, definitionsDir)
			if len(expectedErrs) > 0 {
				// Verify error
				Expect(err).To(HaveOccurred())
				for _, expectedErr := range expectedErrs {
					Expect(err.Error()).To(ContainSubstring(expectedErr))
				}
			} else {
				// Verify no error
				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeEmpty())

				// Render output into unstructured objects
				objs := []client.Object{
					&unstructured.Unstructured{},
					&unstructured.Unstructured{},
					&unstructured.Unstructured{},
				}
				sc.RenderToObjects(objs, output)
				// TODO: make internal render function skip empty documents
				// TODO: add new render helper that returns unstructured objects
				// TODO: use new render helper in both examples (Crossplane and KubeVela)

				// Verify rendered objects
				for _, document := range strings.Split(expectedOutput, "---") {
					Expect(objs).To(ContainElement(sc.MatchYAML(document)))
				}
			}
		},
		Entry("nil annotations", 8080, nil, nil),
		// TODO: add another positive test case
		// TODO: add a negative test case
	)
})

// runVelaDryRun runs `vela dry-run --offline` with given application and definition paths.
func runVelaDryRun(applicationPath, definitionPath string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("vela", "dry-run", "--offline", "-f", applicationPath, "-d", definitionPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run vela dry-run: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
