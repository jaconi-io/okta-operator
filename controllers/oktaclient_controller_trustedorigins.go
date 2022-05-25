package controllers

import (
	"context"
	"fmt"
	oktav1alpha1 "github.com/jaconi-io/okta-operator/api/v1alpha1"
	"github.com/jaconi-io/okta-operator/okta"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	isTrustedOrigin     = okta.IsTrustedOrigin
	createTrustedOrigin = okta.CreateTrustedOrigin
	deleteTrustedOrigin = okta.DeleteTrustedOrigin
)

func updateTrustedOrigins(oktaClient *oktav1alpha1.OktaClient, ctx context.Context) error {
	// Create trusted origins
	log := ctrllog.FromContext(ctx)
	origins := oktaClient.Spec.TrustedOrigins
	for _, origin := range origins {
		isTrustedOrigin, err := isTrustedOrigin(origin)
		log.Info("Queried trusted origin", "origin", origin, "exists", isTrustedOrigin)
		if err != nil {
			return fmt.Errorf("failed to determine if %q is a trusted origin: %w", origin, err)
		}
		if isTrustedOrigin {
			// Nothing to do
			continue
		}
		log.Info("Creating trusted origin", "origin", origin)
		err = createTrustedOrigin(origin)

		if err != nil {
			return fmt.Errorf("failed to create trusted origin %q: %w", origin, err)
		}
	}

	return nil
}

func deleteTrustedOrigins(oktaClient *oktav1alpha1.OktaClient, ctx context.Context) error {
	log := ctrllog.FromContext(ctx)
	origins := oktaClient.Spec.TrustedOrigins
	for _, origin := range origins {
		isTrustedOrigin, err := isTrustedOrigin(origin)
		log.Info("Queried trusted origin", "origin", origin, "exists", isTrustedOrigin)
		if err != nil {
			return fmt.Errorf("failed to determine if %q is a trusted origin: %w", origin, err)
		}
		if isTrustedOrigin {
			log.Info("Deleting trusted origin", "origin", origin)
			err = deleteTrustedOrigin(origin)

			if err != nil {
				return fmt.Errorf("failed to delete trusted origin %q: %w", origin, err)
			}
		}
	}
	return nil
}
