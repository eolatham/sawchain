package sawchain_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: test

type MockTB struct {
	testing.TB
	FailCalled bool
	ErrorLogs  []string
}

func NewMockTB(t testing.TB) *MockTB {
	return &MockTB{TB: t}
}

func (m *MockTB) Failed() bool {
	return m.FailCalled
}

func (m *MockTB) Fail() {
	m.FailCalled = true
}

func (m *MockTB) FailNow() {
	m.FailCalled = true
}

func (m *MockTB) Errorf(format string, args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprintf(format, args...))
}

func (m *MockTB) Error(args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprint(args...))
}

func (m *MockTB) Fatal(args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprint(args...))
	m.FailCalled = true
}

func (m *MockTB) Fatalf(format string, args ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, fmt.Sprintf(format, args...))
	m.FailCalled = true
}

// TODO: expand MockClient to support simulating differnet first N failures for Get, Update, and Delete
type MockClient struct {
	client.Client
	failCreateCount int
	createCallCount int
}

func (m *MockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	m.createCallCount++
	if m.failCreateCount < 0 || m.createCallCount <= m.failCreateCount {
		return fmt.Errorf("simulated create failure")
	}
	return m.Client.Create(ctx, obj, opts...)
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
				// TODO: implement table func
				// - create single-use MockTB
				// - create single-use Sawchain instance with given client and global bindings, but hardcode fast timeout and interval values
				// - pass methodArgs to CreateResourceAndWait directly after passing the global context variable
				// - use expectedErrs to determine if the test should fail or not (and check all failure message substrings)
				// - if should succeed, check that the expectedObject exists with an immediate client Get call
				// - if should succeed and an object was passed in methodArgs, check that the passed object matches expectedObject
			},
			// TODO: add test cases
			// - all possible argument combinations
			// - typed and unstructured objects
			// - all failure cases
			Entry("TODO", testCase{}),
		)
	})
})
