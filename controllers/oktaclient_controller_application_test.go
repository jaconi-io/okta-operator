package controllers

import (
	"testing"
)

func TestUpdateApplicationNotExists(t *testing.T) {
	resetToLocal()

	err := updateApplication(&testAppClient, nil, testRequest, nil)
	if err != nil {
		t.Errorf("error updating application")
	}
	if len(testOktaClients) != 1 {
		t.Errorf("got %d applications, wanted %d", len(testOktaClients), 1)
	}
	if appsCreated != 1 {
		t.Errorf("got %d method calls, wanted %d", appsCreated, 1)
	}
}

func TestUpdateApplicationExists(t *testing.T) {
	resetToLocal()
	_, _ = addTestApplication(testAppClient.Spec.Name, "", nil, nil)

	err := updateApplication(&testAppClient, nil, testRequest, nil)
	if err != nil {
		t.Errorf("error updating application")
	}
	if len(testOktaClients) != 1 {
		t.Errorf("got %d applications, wanted %d", len(testOktaClients), 1)
	}
	if appsCreated != 0 {
		t.Errorf("got %d method calls, wanted %d", appsCreated, 0)
	}
}

func TestDeleteApplication(t *testing.T) {
	resetToLocal()
	_, _ = addTestApplication(testAppClient.Spec.Name, "", nil, nil)

	err := deleteApplication(&testAppClient, nil)
	if err != nil {
		t.Errorf("error deleting application")
	}

	if len(testOktaClients) != 0 {
		t.Errorf("got %d applications, wanted %d", len(testOktaClients), 0)
	}
	if appsDeleted != 1 {
		t.Errorf("got %d method calls, wanted %d", appsDeleted, 1)
	}
}
