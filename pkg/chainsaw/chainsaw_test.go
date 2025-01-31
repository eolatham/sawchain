package chainsaw

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CheckResources", func() {
	var (
		err          error
		templatePath string
		configMap    *corev1.ConfigMap
	)

	BeforeEach(func() {
		// Create a temporary template file
		templateContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  namespace: default
data:
  key: value
`
		templatePath = filepath.Join(GinkgoT().TempDir(), "template.yaml")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		// Delete the test ConfigMap from the cluster
		err = k8sClient.Delete(ctx, configMap)
		Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
	})

	It("should successfully check matching resources", func() {
		// Create a test ConfigMap in the cluster
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Data: map[string]string{
				"key": "value",
			},
		}
		err = k8sClient.Create(ctx, configMap)
		Expect(err).NotTo(HaveOccurred())

		// Test CheckResources
		err = CheckResources(k8sClient, ctx, templatePath, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should fail when resources don't match", func() {
		// Create a test ConfigMap in the cluster with different data
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Data: map[string]string{
				"key": "different-value",
			},
		}
		err := k8sClient.Create(ctx, cm)
		Expect(err).NotTo(HaveOccurred())

		// Test CheckResources
		err = CheckResources(k8sClient, ctx, templatePath, nil)
		Expect(err).To(HaveOccurred())
	})
})
