package controllers

import (
	"testing"
)

func TestUpdateTrustedOriginsAlreadyTrusted(t *testing.T) {
	resetToLocal()
	_ = addTestTrustedOrigin("a")
	_ = addTestTrustedOrigin("b")

	err := updateTrustedOrigins(&testToClient, nil)
	if err != nil {
		t.Errorf("error calling method")
	}
	if len(testTrustedOrigins) != 2 {
		t.Errorf("got %d origins, wanted %d", len(testTrustedOrigins), 2)
	}
	if trustedOriginsCreated != 0 {
		t.Errorf("got %d method calls, wanted %d", trustedOriginsCreated, 0)
	}
}

func TestUpdateTrustedOriginsNotAlreadyTrusted(t *testing.T) {
	resetToLocal()

	err := updateTrustedOrigins(&testToClient, nil)
	if err != nil {
		t.Errorf("error calling method")
	}
	if len(testTrustedOrigins) != 2 {
		t.Errorf("got %d method calls, wanted %d", len(testTrustedOrigins), 2)
	}
	if trustedOriginsCreated != 2 {
		t.Errorf("got %d method calls, wanted %d", trustedOriginsCreated, 2)
	}
}

func TestDeleteTrustedOrigins(t *testing.T) {
	resetToLocal()
	_ = addTestTrustedOrigin("a")
	_ = addTestTrustedOrigin("b")

	err := deleteTrustedOrigins(&testToClient, nil)
	if err != nil {
		t.Errorf("error calling method")
	}
	if len(testTrustedOrigins) != 0 {
		t.Errorf("got %d method calls, wanted %d", len(testTrustedOrigins), 0)
	}
	if trustedOriginsDeleted != 2 {
		t.Errorf("got %d method calls, wanted %d", trustedOriginsDeleted, 2)
	}
}
