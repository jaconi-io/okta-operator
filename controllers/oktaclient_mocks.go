package controllers

import (
	"context"
	"github.com/jaconi-io/okta-operator/api/v1alpha1"
	"github.com/jaconi-io/okta-operator/okta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var testOktaClients = make(map[string]*okta.Application)
var testTrustedOrigins = []string{}
var appsCreated = 0
var trustedOriginsCreated = 0

var testAppClient = v1alpha1.OktaClient{
	TypeMeta:   metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{},
	Spec: v1alpha1.OktaClientSpec{
		Name: "test-client",
	},
	Status: v1alpha1.OktaClientStatus{},
}

var testApp = okta.Application{
	ID:           "xyz",
	ClientID:     "id",
	ClientSecret: "secret",
}

var testRequest = controllerruntime.Request{}

func appCreatorMock(label string, clientUri string, redirectUris []string, postLogoutRedirectUris []string) (*okta.Application, error) {
	appsCreated++
	return addTestApplication(label, clientUri, redirectUris, postLogoutRedirectUris)
}

func addTestApplication(label string, clientUri string, redirectUris []string, postLogoutRedirectUris []string) (*okta.Application, error) {
	testOktaClients[label] = &testApp
	return &testApp, nil
}

func getAppByLabelMock(label string) (*okta.Application, error) {
	return testOktaClients[label], nil
}

func createOrUpdateSecretMock(ctx context.Context, c client.Client, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	return controllerutil.OperationResultNone, nil
}

var testToClient = v1alpha1.OktaClient{
	TypeMeta:   metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{},
	Spec: v1alpha1.OktaClientSpec{
		TrustedOrigins: []string{
			"a", "b",
		},
	},
	Status: v1alpha1.OktaClientStatus{},
}

func addTrustedOriginMock(origin string) error {
	trustedOriginsCreated++
	return addTestTrustedOrigin(origin)
}

func addTestTrustedOrigin(origin string) error {
	testTrustedOrigins = append(testTrustedOrigins, origin)
	return nil
}

func isTrustedMock(origin string) (bool, error) {
	for _, s := range testTrustedOrigins {
		if origin == s {
			return true, nil
		}
	}
	return false, nil
}

func getSecretMock(k8sClient client.Client, ctx context.Context, req controllerruntime.Request, secretName string) error {
	return nil
}

func reset() {
	testOktaClients = make(map[string]*okta.Application)
	testTrustedOrigins = []string{}
	isTrustedOrigin = isTrustedMock
	createTrustedOrigin = addTrustedOriginMock
	getAppByLabel = getAppByLabelMock
	createApp = appCreatorMock
	appsCreated = 0
	trustedOriginsCreated = 0
}

func resetToLocal() {
	reset()
	createOrUpdateSecret = createOrUpdateSecretMock
	getSecret = getSecretMock
}

func resetToCluster() {
	reset()
	createOrUpdateSecret = controllerutil.CreateOrUpdate
	getSecret = getSecretImpl
}
