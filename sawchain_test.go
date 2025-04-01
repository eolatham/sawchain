package sawchain_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain"
	"github.com/eolatham/sawchain/internal/util"
)

// MockT allows capturing failures and error logs.
type MockT struct {
	testing.TB
	failed    bool
	ErrorLogs []string
}

func NewMockT(t testing.TB) *MockT {
	return &MockT{TB: t}
}

func (m *MockT) Failed() bool {
	return m.failed
}

func (m *MockT) Fail() {
	m.failed = true
}

func (m *MockT) FailNow() {
	m.failed = true
	runtime.Goexit()
}

func (m *MockT) Errorf(format string, args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprintf(format, args...))
}

func (m *MockT) Error(args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprint(args...))
}

func (m *MockT) Fatal(args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprint(args...))
	m.failed = true
}

func (m *MockT) Fatalf(format string, args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprintf(format, args...))
	m.failed = true
	runtime.Goexit()
}

// MockClient allows simulating K8s API failures.
type MockClient struct {
	client.Client

	getFailFirstN int
	getCallCount  int

	createFailFirstN int
	createCallCount  int

	updateFailFirstN int
	updateCallCount  int

	deleteFailFirstN int
	deleteCallCount  int
}

func NewMockClient(c client.Client) *MockClient {
	return &MockClient{Client: c}
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	m.getCallCount++
	if m.getFailFirstN < 0 || m.getCallCount <= m.getFailFirstN {
		return fmt.Errorf("simulated get failure")
	}
	return m.Client.Get(ctx, key, obj, opts...)
}

func (m *MockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	m.createCallCount++
	if m.createFailFirstN < 0 || m.createCallCount <= m.createFailFirstN {
		return fmt.Errorf("simulated create failure")
	}
	return m.Client.Create(ctx, obj, opts...)
}

func (m *MockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	m.updateCallCount++
	if m.updateFailFirstN < 0 || m.updateCallCount <= m.updateFailFirstN {
		return fmt.Errorf("simulated update failure")
	}
	return m.Client.Update(ctx, obj, opts...)
}

func (m *MockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	m.deleteCallCount++
	if m.deleteFailFirstN < 0 || m.deleteCallCount <= m.deleteFailFirstN {
		return fmt.Errorf("simulated delete failure")
	}
	return m.Client.Delete(ctx, obj, opts...)
}

var _ = Describe("Sawchain", func() {
	Describe("CreateResourceAndWait", func() {
		type testCase struct {
			client         client.Client
			globalBindings map[string]any
			methodArgs     []interface{}
			expectedErrs   []string
			expectedObject client.Object
		}
		DescribeTable("creating a test resource",
			func(tc testCase) {
				// Create Sawchain
				t := NewMockT(GinkgoTB())
				sc := sawchain.New(t, tc.client, fastTimeout, fastInterval, tc.globalBindings)

				// Test CreateResourceAndWait
				done := make(chan struct{})
				go func() {
					defer close(done)
					sc.CreateResourceAndWait(ctx, tc.methodArgs...)
				}()
				<-done

				if len(tc.expectedErrs) > 0 {
					// Verify failure
					Expect(t.Failed()).To(BeTrue(), "expected CreateResourceAndWait to fail")
					for _, expectedErr := range tc.expectedErrs {
						Expect(t.ErrorLogs).To(ContainElement(ContainSubstring(expectedErr)))
					}
				} else {
					// Verify successful creation
					Expect(tc.client.Get(ctx, client.ObjectKeyFromObject(tc.expectedObject), tc.expectedObject)).To(Succeed(), "expected CreateResourceAndWait to create resource")
					// Verify resource state
					for _, arg := range tc.methodArgs {
						if obj, ok := util.AsObject(arg); ok {
							Expect(obj).To(Equal(tc.expectedObject), "expected CreateResourceAndWait to save created resource state to provided object")
							break
						}
					}
				}
			},
			Entry("should fail with no arguments", testCase{
				client:         NewMockClient(standardClient),
				globalBindings: map[string]any{},
				methodArgs:     []interface{}{},
				expectedErrs: []string{
					"invalid arguments",
					"required argument(s) not provided: Template (string) or Object (client.Object)",
				},
			}),
			// TODO: add test cases
			// - all possible argument combinations
			// - typed and unstructured objects
			// - all failure cases
		)
	})
})
