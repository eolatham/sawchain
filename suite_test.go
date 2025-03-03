package sawchain_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	ctx       context.Context
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func TestSawchain(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(time.Second * 10)
	SetDefaultEventuallyPollingInterval(time.Second * 1)
	SetDefaultConsistentlyDuration(time.Second * 5)
	SetDefaultConsistentlyPollingInterval(time.Second * 1)
	suiteConfig := ginkgotypes.SuiteConfig{
		Timeout:         time.Minute * 5,
		GracePeriod:     time.Second * 10,
		ParallelTotal:   1,
		ParallelProcess: 1,
		ParallelHost:    "N/A",
	}
	RunSpecs(t, "Sawchain Suite", suiteConfig)
}

var _ = BeforeSuite(func() {
	ctrl.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx = context.Background()

	testEnv = &envtest.Environment{}

	config, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(config).NotTo(BeNil())

	k8sClient, err = client.New(config, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	Eventually(testEnv.Stop, time.Second*30, time.Second*2).Should(Succeed())
})
