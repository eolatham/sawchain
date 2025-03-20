package testutil_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/eolatham/sawchain/internal/testutil"
)

var _ = Describe("Testutil", func() {
	DescribeTable("CreateTempDir",
		func(namePattern string) {
			tempDirPath := testutil.CreateTempDir(namePattern)

			// Verify the directory exists and has the right pattern
			Expect(tempDirPath).To(ContainSubstring(namePattern))
			info, err := os.Stat(tempDirPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())

			// Clean up
			os.RemoveAll(tempDirPath)
		},
		Entry("with test pattern", "test-pattern"),
		Entry("with empty pattern", ""),
	)

	DescribeTable("CreateTempFile",
		func(namePattern, content string) {
			tempFilePath := testutil.CreateTempFile(namePattern, content)

			// Verify the file exists and has the right pattern and content
			Expect(tempFilePath).To(ContainSubstring(namePattern))
			fileContent, err := os.ReadFile(tempFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(fileContent)).To(Equal(content))

			// Clean up
			os.Remove(tempFilePath)
		},
		Entry("with test pattern and content", "test-file-pattern", "test content"),
		Entry("with empty pattern and content", "", "test content"),
		Entry("with pattern and empty content", "test-file-pattern", ""),
	)

	DescribeTable("CreateEmptyScheme",
		func() {
			scheme := testutil.CreateEmptyScheme()
			Expect(scheme).NotTo(BeNil())
			// Empty scheme should not have any types registered
			Expect(scheme.AllKnownTypes()).To(HaveLen(0))
		},
		Entry("creates empty scheme"),
	)

	DescribeTable("CreateStandardScheme",
		func() {
			scheme := testutil.CreateStandardScheme()
			Expect(scheme).NotTo(BeNil())
			// Standard scheme should have types registered from k8s
			Expect(scheme.AllKnownTypes()).NotTo(BeEmpty())
		},
		Entry("creates standard scheme"),
	)
})
