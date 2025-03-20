package testutil

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// CreateTempDir creates a temporary directory and returns its path.
func CreateTempDir(namePattern string) string {
	tempDir, err := os.MkdirTemp("", namePattern)
	if err != nil {
		panic(err)
	}
	return tempDir
}

// CreateTempFile creates a temporary file and returns its path.
func CreateTempFile(namePattern, content string) string {
	file, err := os.CreateTemp("", namePattern)
	if err != nil {
		panic(err)
	}
	path := file.Name()
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	return path
}

// CreateEmptyScheme returns a new empty runtime.Scheme.
func CreateEmptyScheme() *runtime.Scheme {
	return runtime.NewScheme()
}

// CreateStandardScheme returns a new standard runtime.scheme supporting built-in APIs.
func CreateStandardScheme() *runtime.Scheme {
	s := CreateEmptyScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		panic(err)
	}
	return s
}
