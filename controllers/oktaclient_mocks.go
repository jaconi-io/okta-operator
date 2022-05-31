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

type stringSlice []string

var testOktaClients = make(map[string]*okta.Application)
var testTrustedOrigins = stringSlice{}
var appsCreated = 0
var appsDeleted = 0
var trustedOriginsCreated = 0
var trustedOriginsDeleted = 0

var testAppClient = v1alpha1.OktaClient{
	TypeMeta:   metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{},
	Spec: v1alpha1.OktaClientSpec{
		Name: "test-client",
	},
	Status: v1alpha1.OktaClientStatus{},
}

var testApp = okta.Application{
	ID:           "test-client",
	ClientID:     "id",
	ClientSecret: "secret",
}

var testRequest = controllerruntime.Request{}

func deleteAppMock(app *okta.Application) error {
	appsDeleted++
	delete(testOktaClients, app.ID)
	return nil
}

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

func newSecretMock(clientID string) (string, error) {
	return "secret", nil
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

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func pos(s []string, value string) int {
	for p, v := range s {
		if v == value {
			return p
		}
	}
	return -1
}

func deleteTrustedOriginMock(origin string) error {
	trustedOriginsDeleted++
	pos := pos(testTrustedOrigins, origin)
	if pos >= 0 {
		testTrustedOrigins = removeIndex(testTrustedOrigins, pos)
	}
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
	deleteTrustedOrigin = deleteTrustedOriginMock
	getAppByLabel = getAppByLabelMock
	deleteApp = deleteAppMock
	createApp = appCreatorMock
	newSecret = newSecretMock
	appsCreated = 0
	appsDeleted = 0
	trustedOriginsCreated = 0
	trustedOriginsDeleted = 0
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
