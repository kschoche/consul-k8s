package catalog

import (
	"testing"
	"time"

	"github.com/hashicorp/consul-k8s/helper/controller"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func init() {
	hclog.DefaultOptions.Level = hclog.Debug
}

func TestServiceResource_impl(t *testing.T) {
	var _ controller.Resource = &ServiceResource{}
	var _ controller.Backgrounder = &ServiceResource{}
}

// Test that deleting a service properly deletes the registration.
func TestServiceResource_createDelete(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	client := fake.NewSimpleClientset()
	syncer := &TestSyncer{}

	// Start the controller
	closer := controller.TestControllerRun(&ServiceResource{
		Log:    hclog.Default(),
		Client: client,
		Syncer: syncer,
	})
	defer closer()

	// Insert an LB service
	_, err := client.CoreV1().Services(metav1.NamespaceDefault).Create(testService("foo"))
	require.NoError(err)
	time.Sleep(100 * time.Millisecond)

	// Delete
	require.NoError(client.CoreV1().Services(metav1.NamespaceDefault).Delete("foo", nil))
	time.Sleep(300 * time.Millisecond)

	// Verify what we got
	syncer.Lock()
	defer syncer.Unlock()
	actual := syncer.Registrations
	require.Len(actual, 0)
}

// Test that the proper registrations are generated for a LoadBalancer.
func TestServiceResource_lb(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	client := fake.NewSimpleClientset()
	syncer := &TestSyncer{}

	// Start the controller
	closer := controller.TestControllerRun(&ServiceResource{
		Log:    hclog.Default(),
		Client: client,
		Syncer: syncer,
	})
	defer closer()

	// Insert an LB service
	_, err := client.CoreV1().Services(metav1.NamespaceDefault).Create(&apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},

		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},

		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					apiv1.LoadBalancerIngress{
						IP: "1.2.3.4",
					},
				},
			},
		},
	})
	require.NoError(err)

	// Wait a bit
	time.Sleep(300 * time.Millisecond)

	// Verify what we got
	syncer.Lock()
	defer syncer.Unlock()
	actual := syncer.Registrations
	require.Len(actual, 1)
	require.Equal("foo", actual[0].Service.Service)
	require.Equal("1.2.3.4", actual[0].Service.Address)
}

// testService returns a service that will result in a registration.
func testService(name string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},

		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},

		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					apiv1.LoadBalancerIngress{
						IP: "1.2.3.4",
					},
				},
			},
		},
	}
}