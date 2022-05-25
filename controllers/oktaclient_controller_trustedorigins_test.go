package controllers

import (
	"github.com/jaconi-io/okta-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var toCount = 0

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

func countingToCreator(origin string) error {
	toCount++
	return nil
}

func trustedMock(origin string) (bool, error) {
	return true, nil
}

func notTrustedMock(origin string) (bool, error) {
	return false, nil
}

func TestUpdateTrustedOriginsAlreadyTrusted(t *testing.T) {
	toCount = 0
	isTrustedOrigin = trustedMock
	createTrustedOrigin = countingToCreator

	err := updateTrustedOrigins(&testToClient, nil)
	if err != nil {
		t.Errorf("error calling method")
	}
	if toCount != 0 {
		t.Errorf("got %d method calls, wanted %d", toCount, 0)
	}
}

func TestUpdateTrustedOriginsNotAlreadyTrusted(t *testing.T) {
	toCount = 0
	isTrustedOrigin = notTrustedMock
	createTrustedOrigin = countingToCreator

	err := updateTrustedOrigins(&testToClient, nil)
	if err != nil {
		t.Errorf("error calling method")
	}
	if toCount != 2 {
		t.Errorf("got %d method calls, wanted %d", toCount, 2)
	}
}
