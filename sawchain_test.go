package sawchain_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain"
	"github.com/eolatham/sawchain/internal/testutil"
	"github.com/eolatham/sawchain/internal/util"
)

// MockT allows capturing failures and error logs.
type MockT struct {
	testing.TB
	failed    bool
	ErrorLogs []string
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
			client           client.Client
			globalBindings   map[string]any
			methodArgs       []interface{}
			expectedErrs     []string
			expectedObject   client.Object
			expectedDuration time.Duration
		}
		DescribeTable("creating a test resource",
			func(tc testCase) {
				// Create Sawchain
				t := &MockT{TB: GinkgoTB()}
				sc := sawchain.New(t, tc.client, fastTimeout, fastInterval, tc.globalBindings)

				// Test CreateResourceAndWait
				done := make(chan struct{})
				start := time.Now()
				go func() {
					defer close(done)
					sc.CreateResourceAndWait(ctx, tc.methodArgs...)
				}()
				<-done
				executionTime := time.Since(start)

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

				// Verify execution time
				if tc.expectedDuration > 0 {
					maxAllowedDuration := time.Duration(float64(tc.expectedDuration) * 1.1)
					Expect(executionTime).To(BeNumerically("<", maxAllowedDuration),
						"expected CreateResourceAndWait to complete in less than %v, but took %v",
						maxAllowedDuration, executionTime)
				}
			},

			// Success cases
			Entry("should create ConfigMap with typed object",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					},
					expectedObject:   testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create ConfigMap with unstructured object",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewUnstructuredConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					},
					expectedObject:   testutil.NewUnstructuredConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create custom resource with typed object",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClientWithTestResource()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewTestResource("test-cr", "default", []metav1.Condition{}),
					},
					expectedObject:   testutil.NewTestResource("test-cr", "default", []metav1.Condition{}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create custom resource with unstructured object",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClientWithTestResource()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewUnstructuredTestResource("test-cr", "default", []metav1.Condition{}),
					},
					expectedObject:   testutil.NewUnstructuredTestResource("test-cr", "default", []metav1.Condition{}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create ConfigMap with static template string",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  key: value
`,
					},
					expectedObject:   testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create ConfigMap with template string and bindings",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{"namespace": "test-ns"},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($name)
  namespace: ($namespace)
data:
  key: ($value)
`,
						map[string]any{"name": "test-cm", "value": "configured-value"},
					},
					expectedObject:   testutil.NewConfigMap("test-cm", "test-ns", map[string]string{"key": "configured-value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create ConfigMap with template string and multiple binding maps",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{"namespace": "test-ns", "name": "test-cm"},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($name)
  namespace: ($namespace)
data:
  key: ($value)
`,
						map[string]any{"name": "override-cm", "value": "first-value"},
						map[string]any{"value": "override-value"},
					},
					expectedObject:   testutil.NewConfigMap("override-cm", "test-ns", map[string]string{"key": "override-value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create ConfigMap with template string and save to typed object",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewConfigMap("placeholder", "placeholder", nil),
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  key: value
`,
					},
					expectedObject:   testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create ConfigMap with template string with bindings and save to typed object",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{"namespace": "test-ns"},
					methodArgs: []interface{}{
						testutil.NewConfigMap("placeholder", "placeholder", nil),
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($name)
  namespace: ($namespace)
data:
  key: ($value)
`,
						map[string]any{"name": "test-cm", "value": "configured-value"},
					},
					expectedObject:   testutil.NewConfigMap("test-cm", "test-ns", map[string]string{"key": "configured-value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create ConfigMap with template string and save to unstructured object",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						&unstructured.Unstructured{},
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  key: value
`,
					},
					expectedObject:   testutil.NewUnstructuredConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					expectedDuration: fastTimeout,
				},
			),

			Entry("should respect custom timeout and interval",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
						"50ms", // Custom timeout
						"10ms", // Custom interval
					},
					expectedObject:   testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					expectedDuration: 50 * time.Millisecond,
				},
			),

			Entry("should handle transient get failures",
				testCase{
					client: &MockClient{
						Client:        testutil.NewStandardFakeClient(),
						getFailFirstN: 2, // Fail the first 2 get attempts
					},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					},
					expectedObject:   testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					expectedDuration: fastTimeout,
				},
			),

			// Failure cases
			Entry("should fail with no arguments",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs:     []interface{}{},
					expectedErrs: []string{
						"invalid arguments",
						"required argument(s) not provided: Template (string) or Object (client.Object)",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail with unexpected argument type",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]string{"unexpected", "argument", "type"},
					},
					expectedErrs: []string{
						"invalid arguments",
						"unexpected argument type: []string",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail with invalid template",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						`invalid: yaml: [`,
					},
					expectedErrs: []string{
						"invalid template/bindings",
						"failed to parse template",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail with template binding errors",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($missing)
  namespace: default
`,
					},
					expectedErrs: []string{
						"invalid template/bindings",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail when create fails",
				testCase{
					client: &MockClient{
						Client:           testutil.NewStandardFakeClient(),
						createFailFirstN: 1,
					},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					},
					expectedErrs: []string{
						"failed to create with object",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail when get fails indefinitely after create",
				testCase{
					client: &MockClient{
						Client:        testutil.NewStandardFakeClient(),
						getFailFirstN: -1, // Fail all get attempts
					},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						testutil.NewConfigMap("test-cm", "default", map[string]string{"key": "value"}),
					},
					expectedErrs: []string{
						"client cache not synced within timeout",
					},
					expectedDuration: fastTimeout,
				},
			),
		)
	})

	Describe("CreateResourcesAndWait", func() {
		type testCase struct {
			client           client.Client
			globalBindings   map[string]any
			methodArgs       []interface{}
			expectedErrs     []string
			expectedObjects  []client.Object
			expectedDuration time.Duration
		}
		DescribeTable("creating multiple test resources",
			func(tc testCase) {
				// Create Sawchain
				t := &MockT{TB: GinkgoTB()}
				sc := sawchain.New(t, tc.client, fastTimeout, fastInterval, tc.globalBindings)

				// Test CreateResourcesAndWait
				done := make(chan struct{})
				start := time.Now()
				go func() {
					defer close(done)
					sc.CreateResourcesAndWait(ctx, tc.methodArgs...)
				}()
				<-done
				executionTime := time.Since(start)

				if len(tc.expectedErrs) > 0 {
					// Verify failure
					Expect(t.Failed()).To(BeTrue(), "expected CreateResourcesAndWait to fail")
					for _, expectedErr := range tc.expectedErrs {
						Expect(t.ErrorLogs).To(ContainElement(ContainSubstring(expectedErr)))
					}
				} else {
					// Verify successful creation of all resources
					for _, expectedObject := range tc.expectedObjects {
						Expect(tc.client.Get(ctx, client.ObjectKeyFromObject(expectedObject), expectedObject)).To(Succeed(),
							"expected CreateResourcesAndWait to create resource: %s", client.ObjectKeyFromObject(expectedObject))
					}

					// Verify resource states in provided objects slice
					for _, arg := range tc.methodArgs {
						if objects, ok := arg.([]client.Object); ok {
							Expect(objects).To(HaveLen(len(tc.expectedObjects)), "expected objects slice to have the same length as expected objects")
							for i, obj := range objects {
								Expect(obj).To(Equal(tc.expectedObjects[i]), "expected CreateResourcesAndWait to save created resource state to provided object")
							}
							break
						}
					}
				}

				// Verify execution time
				if tc.expectedDuration > 0 {
					maxAllowedDuration := time.Duration(float64(tc.expectedDuration) * 1.1)
					Expect(executionTime).To(BeNumerically("<", maxAllowedDuration),
						"expected CreateResourcesAndWait to complete in less than %v, but took %v",
						maxAllowedDuration, executionTime)
				}
			},

			// Success cases
			Entry("should create multiple resources with typed objects",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
							testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
						},
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple resources with unstructured objects",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewUnstructuredConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
							testutil.NewUnstructuredConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
						},
					},
					expectedObjects: []client.Object{
						testutil.NewUnstructuredConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						testutil.NewUnstructuredConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple custom resources with typed objects",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClientWithTestResource()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewTestResource("test-cr1", "default", []metav1.Condition{}),
							testutil.NewTestResource("test-cr2", "default", []metav1.Condition{}),
						},
					},
					expectedObjects: []client.Object{
						testutil.NewTestResource("test-cr1", "default", []metav1.Condition{}),
						testutil.NewTestResource("test-cr2", "default", []metav1.Condition{}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple custom resources with unstructured objects",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClientWithTestResource()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewUnstructuredTestResource("test-cr1", "default", []metav1.Condition{}),
							testutil.NewUnstructuredTestResource("test-cr2", "default", []metav1.Condition{}),
						},
					},
					expectedObjects: []client.Object{
						testutil.NewUnstructuredTestResource("test-cr1", "default", []metav1.Condition{}),
						testutil.NewUnstructuredTestResource("test-cr2", "default", []metav1.Condition{}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple resources with static template string",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm1
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
  namespace: default
data:
  key2: value2
`,
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple resources with template string and bindings",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{"namespace": "test-ns"},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: (join('-', [$prefix, 'cm1']))
  namespace: ($namespace)
data:
  key1: ($value1)
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: (join('-', [$prefix, 'cm2']))
  namespace: ($namespace)
data:
  key2: ($value2)
`,
						map[string]any{
							"prefix": "test",
							"value1": "configured-value1",
							"value2": "configured-value2",
						},
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("test-cm1", "test-ns", map[string]string{"key1": "configured-value1"}),
						testutil.NewConfigMap("test-cm2", "test-ns", map[string]string{"key2": "configured-value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple resources with template string and multiple binding maps",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{"namespace": "test-ns", "prefix": "global"},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: (join('-', [$prefix, 'cm1']))
  namespace: ($namespace)
data:
  key1: ($value1)
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: (join('-', [$prefix, 'cm2']))
  namespace: ($namespace)
data:
  key2: ($value2)
`,
						map[string]any{"prefix": "local", "value1": "first-value"},
						map[string]any{"value1": "override1", "value2": "override2"},
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("local-cm1", "test-ns", map[string]string{"key1": "override1"}),
						testutil.NewConfigMap("local-cm2", "test-ns", map[string]string{"key2": "override2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple resources with template string and save to typed objects",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("placeholder1", "placeholder", nil),
							testutil.NewConfigMap("placeholder2", "placeholder", nil),
						},
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm1
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
  namespace: default
data:
  key2: value2
`,
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple resources with template string with bindings and save to typed objects",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{"namespace": "test-ns"},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("placeholder1", "placeholder", nil),
							testutil.NewConfigMap("placeholder2", "placeholder", nil),
						},
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: (join('-', [$prefix, 'cm1']))
  namespace: ($namespace)
data:
  key1: ($value1)
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: (join('-', [$prefix, 'cm2']))
  namespace: ($namespace)
data:
  key2: ($value2)
`,
						map[string]any{
							"prefix": "test",
							"value1": "configured-value1",
							"value2": "configured-value2",
						},
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("test-cm1", "test-ns", map[string]string{"key1": "configured-value1"}),
						testutil.NewConfigMap("test-cm2", "test-ns", map[string]string{"key2": "configured-value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should create multiple resources with template string and save to unstructured objects",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							&unstructured.Unstructured{},
							&unstructured.Unstructured{},
						},
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm1
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
  namespace: default
data:
  key2: value2
`,
					},
					expectedObjects: []client.Object{
						testutil.NewUnstructuredConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						testutil.NewUnstructuredConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should respect custom timeout and interval",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
							testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
						},
						"50ms", // Custom timeout
						"10ms", // Custom interval
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
					},
					expectedDuration: 50 * time.Millisecond,
				},
			),

			Entry("should handle transient get failures",
				testCase{
					client: &MockClient{
						Client:        testutil.NewStandardFakeClient(),
						getFailFirstN: 2, // Fail the first 2 get attempts
					},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
							testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
						},
					},
					expectedObjects: []client.Object{
						testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
					},
					expectedDuration: fastTimeout,
				},
			),

			// Failure cases
			Entry("should fail with no arguments",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs:     []interface{}{},
					expectedErrs: []string{
						"invalid arguments",
						"required argument(s) not provided: Template (string) or Objects ([]client.Object)",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail with unexpected argument type",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						42, // Unexpected argument type
					},
					expectedErrs: []string{
						"invalid arguments",
						"unexpected argument type: int",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail with invalid template",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						`invalid: yaml: [`,
					},
					expectedErrs: []string{
						"invalid template/bindings",
						"failed to parse template",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail with template binding errors",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($missing)
  namespace: default
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
  namespace: ($also-missing)
`,
					},
					expectedErrs: []string{
						"invalid template/bindings",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail when create fails",
				testCase{
					client: &MockClient{
						Client:           testutil.NewStandardFakeClient(),
						createFailFirstN: 1,
					},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
							testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
						},
					},
					expectedErrs: []string{
						"failed to create with object",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail with object length mismatch",
				testCase{
					client:         &MockClient{Client: testutil.NewStandardFakeClient()},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
						},
						`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm1
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm2
  namespace: default
data:
  key2: value2
`,
					},
					expectedErrs: []string{
						"invalid objects slice: length must match template resource count",
					},
					expectedDuration: fastTimeout,
				},
			),

			Entry("should fail when get fails indefinitely after create",
				testCase{
					client: &MockClient{
						Client:        testutil.NewStandardFakeClient(),
						getFailFirstN: -1, // Fail all get attempts
					},
					globalBindings: map[string]any{},
					methodArgs: []interface{}{
						[]client.Object{
							testutil.NewConfigMap("test-cm1", "default", map[string]string{"key1": "value1"}),
							testutil.NewConfigMap("test-cm2", "default", map[string]string{"key2": "value2"}),
						},
					},
					expectedErrs: []string{
						"client cache not synced within timeout",
					},
					expectedDuration: fastTimeout,
				},
			),
		)
	})
})
