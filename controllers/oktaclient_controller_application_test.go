package controllers

import (
	"context"
	"github.com/jaconi-io/okta-operator/api/v1alpha1"
	"github.com/jaconi-io/okta-operator/okta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"testing"
)

var appCount = 0

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

func countingAppCreator(label string, clientUri string, redirectUris []string, postLogoutRedirectUris []string) (*okta.Application, error) {
	appCount++
	return &testApp, nil
}

func getAppByLabelNotExistsMock(label string) (*okta.Application, error) {
	return nil, nil
}

func getAppByLabelExistsMock(label string) (*okta.Application, error) {
	return &okta.Application{}, nil
}

func createOrUpdateSecretMock(ctx context.Context, c client.Client, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	return controllerutil.OperationResultNone, nil
}

func TestUpdateApplicationNotExists(t *testing.T) {
	appCount = 0
	getAppByLabel = getAppByLabelNotExistsMock
	createApp = countingAppCreator
	createOrUpdateSecret = createOrUpdateSecretMock

	err := updateApplication(&testAppClient, nil, testRequest, nil)
	if err != nil {
		t.Errorf("error updating application")
	}
	if appCount != 1 {
		t.Errorf("got %d method calls, wanted %d", appCount, 1)
	}
}

func getSecretMock(k8sClient client.Client, ctx context.Context, req controllerruntime.Request, secretName string) error {
	return nil
}

func TestUpdateApplicationExists(t *testing.T) {
	appCount = 0
	getAppByLabel = getAppByLabelExistsMock
	createApp = countingAppCreator
	createOrUpdateSecret = createOrUpdateSecretMock
	getSecret = getSecretMock

	err := updateApplication(&testAppClient, nil, testRequest, nil)
	if err != nil {
		t.Errorf("error updating application")
	}
	if appCount != 0 {
		t.Errorf("got %d method calls, wanted %d", appCount, 0)
	}
}
