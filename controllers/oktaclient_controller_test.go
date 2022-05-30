package controllers

import (
	"context"
	"github.com/jaconi-io/okta-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("OktaClient controller", func() {
	const (
		OktaClientName      = "test-oktaclient"
		OktaClientNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	BeforeEach(func() {
		resetToCluster()
		getAppByLabel = getAppByLabelMock
		createApp = appCreatorMock
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("When creating an OktaClient And the Okta Application does not exist", func() {
		It("Should Create the Application and a Secret", func() {
			ctx := context.Background()
			oktaClient := &v1alpha1.OktaClient{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "okta.jaconi.io/v1alpha1",
					Kind:       "OktaClient",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      OktaClientName,
					Namespace: OktaClientNamespace,
				},
				Spec: v1alpha1.OktaClientSpec{
					Name: OktaClientName,
				},
				Status: v1alpha1.OktaClientStatus{},
			}
			Expect(k8sClient.Create(ctx, oktaClient)).Should(Succeed())

			secretLookupKey := types.NamespacedName{Name: OktaClientName, Namespace: OktaClientNamespace}
			createdSecret := &core.Secret{}

			// App created
			Eventually(func() int {
				return len(testOktaClients)
			}, timeout, interval).Should(Equal(1))

			// We'll need to retry getting this newly created Secret, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, createdSecret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

		})
	})
})
