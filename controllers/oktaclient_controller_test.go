package controllers

import (
	"context"
	"github.com/jaconi-io/okta-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("OktaClient controller", func() {
	const (
		OktaClientName = "test-client"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	ns := &core.Namespace{}

	BeforeEach(func() {
		resetToCluster()
		getAppByLabel = getAppByLabelMock
		createApp = appCreatorMock

		*ns = core.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "testns-" + randStringRunes(5)},
		}

		err := k8sClient.Create(ctx, ns)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
	})

	AfterEach(func() {
		deleteOptions := ctrl.DeleteOptions{}
		gracePeriodSeconds := int64(0)
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
		err := k8sClient.Delete(ctx, ns, &deleteOptions)
		Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
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
					Namespace: ns.Name,
				},
				Spec: v1alpha1.OktaClientSpec{
					Name: OktaClientName,
					TrustedOrigins: []string{
						"a", "b",
					},
				},
				Status: v1alpha1.OktaClientStatus{},
			}
			Expect(k8sClient.Create(ctx, oktaClient)).Should(Succeed())

			secretLookupKey := types.NamespacedName{Name: OktaClientName, Namespace: ns.Name}
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

	Context("When deleting an OktaClient", func() {
		It("Should Delete the Application, but not the Secret", func() {
			ctx := context.Background()
			oktaClient := &v1alpha1.OktaClient{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "okta.jaconi.io/v1alpha1",
					Kind:       "OktaClient",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      OktaClientName,
					Namespace: ns.Name,
				},
				Spec: v1alpha1.OktaClientSpec{
					Name: OktaClientName,
					TrustedOrigins: []string{
						"a", "b",
					},
				},
				Status: v1alpha1.OktaClientStatus{},
			}

			Expect(k8sClient.Create(ctx, oktaClient)).Should(Succeed())

			secretLookupKey := types.NamespacedName{Name: OktaClientName, Namespace: ns.Name}
			createdSecret := &core.Secret{}

			// App created
			Eventually(func() int {
				return len(testOktaClients)
			}, timeout, interval).Should(Equal(1))

			// Trusted Origins created
			Eventually(func() int {
				return len(testTrustedOrigins)
			}, timeout, interval).Should(Equal(2))

			// We'll need to retry getting this newly created Secret, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, createdSecret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, oktaClient)).Should(Succeed())

			// App deleted
			Eventually(func() int {
				return len(testOktaClients)
			}, timeout, interval).Should(Equal(0))

			// Trusted Origins deleted
			Eventually(func() int {
				return len(testTrustedOrigins)
			}, timeout, interval).Should(Equal(0))

			// Created secret stays there
			Consistently(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, createdSecret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

		})
	})
})
