package chainsaw_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	ctx       context.Context
	testEnv   *envtest.Environment
	k8sClient client.Client
)

func TestChainsaw(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(time.Second * 10)
	SetDefaultEventuallyPollingInterval(time.Second * 1)
	SetDefaultConsistentlyDuration(time.Second * 5)
	SetDefaultConsistentlyPollingInterval(time.Second * 1)
	suiteConfig := ginkgotypes.SuiteConfig{
		Timeout:         time.Minute * 1,
		GracePeriod:     time.Second * 10,
		ParallelTotal:   1,
		ParallelProcess: 1,
		ParallelHost:    "N/A",
	}
	RunSpecs(t, "Chainsaw Suite", suiteConfig)
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

	// Create namespace used in tests
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "other-namespace",
		},
	}
	Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "Failed to create other-namespace")
	Eventually(func() error {
		return k8sClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)
	}).Should(Succeed(), "Timed out waiting for other-namespace to be created")
})

var _ = AfterSuite(func() {
	Eventually(testEnv.Stop, time.Second*30, time.Second*2).Should(Succeed())
})
