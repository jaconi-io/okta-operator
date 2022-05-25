package controllers

import (
	"context"
	"fmt"
	oktav1alpha1 "github.com/jaconi-io/okta-operator/api/v1alpha1"
	"github.com/jaconi-io/okta-operator/okta"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	getAppByLabel         = okta.GetApplicationByLabel
	createApp             = okta.CreateApplication
	deleteApp             = okta.DeleteApplication
	newSecret             = okta.NewSecret
	createGroupAssignment = okta.CreateApplicationGroupAssignment
	createOrUpdateSecret  = controllerutil.CreateOrUpdate
	getSecret             = getSecretImpl
)

func updateApplication(oktaClient *oktav1alpha1.OktaClient, ctx context.Context, req ctrl.Request, k8sClient client.Client) error {
	// Update application
	log := ctrllog.FromContext(ctx)
	secretName := oktaClient.Name
	appName := oktaClient.Spec.Name
	clientUri := oktaClient.Spec.ClientUri
	redirectUris := oktaClient.Spec.RedirectUris
	postLogoutRedirectUris := oktaClient.Spec.PostLogoutRedirectUris
	groupId := oktaClient.Spec.GroupId

	app, err := getAppByLabel(appName)
	log.Info("Queried application", "application", appName, "exists", app != nil)
	if err != nil {
		return fmt.Errorf("failed to get application %q: %w", appName, err)
	}

	if app == nil {
		log.Info("Creating application", "application", appName)
		app, err = createApp(appName, clientUri, redirectUris, postLogoutRedirectUris)
		if err != nil {
			return fmt.Errorf("failed to create application %q: %w", appName, err)
		}
	} else {
		// The application has already been created in Okta. Check if we have the client credentials for the application.
		err := getSecret(k8sClient, ctx, req, secretName)
		if err != nil {
			if errors.IsNotFound(err) {
				// The secret does not exist, and we do not have the credentials at hand. Create a new secret.
				log.Info("Rotating application secret")
				clientSecret, err := newSecret(app.ClientID)
				if err != nil {
					return fmt.Errorf("could not rotate secret for application %q: %w", appName, err)
				}
				app.ClientSecret = clientSecret
			}
		}
	}

	// If we have a new ClientSecret, create or update the K8s secret
	if app.ClientSecret != "" {
		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: req.Namespace,
			},
		}
		_, err = createOrUpdateSecret(ctx, k8sClient, secret, func() error {
			secret.StringData = map[string]string{
				"OKTA_CLIENT_ID":     app.ClientID,
				"OKTA_CLIENT_SECRET": app.ClientSecret,
			}

			return nil
		})
	}
	if err != nil {
		return fmt.Errorf("failed to create / update secret for application %q: %w", appName, err)
	}

	if groupId != "" {
		log.Info("Creating application/group assignment", "application", appName, "groupId", groupId)
		err = createGroupAssignment(app, groupId)
		if err != nil {
			return fmt.Errorf("failed to add application %q to group %q: %w", appName, groupId, err)
		}
	}

	return nil
}

func getSecretImpl(k8sClient client.Client, ctx context.Context, req ctrl.Request, secretName string) error {
	return k8sClient.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: secretName}, &core.Secret{})
}

func deleteApplication(oktaClient *oktav1alpha1.OktaClient, ctx context.Context) error {
	log := ctrllog.FromContext(ctx)
	appName := oktaClient.Spec.Name
	app, err := getAppByLabel(appName)

	log.Info("Queried application", "appName", appName, "exists", app != nil)
	if err != nil {
		return fmt.Errorf("failed to get application %q: %w", appName, err)
	}

	if app != nil {
		log.Info("Deleting application", "appName", appName)
		err = deleteApp(app)
		if err != nil {
			return fmt.Errorf("failed to delete application %q: %w", appName, err)
		}
	}
	return nil
}
