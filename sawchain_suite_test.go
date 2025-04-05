package sawchain_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	fastTimeout  = 100 * time.Millisecond
	fastInterval = 5 * time.Millisecond
)

// Variables must be assigned inline to beat static Entry parsing!
var ctx = context.Background()

func TestSawchain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sawchain Suite")
}
