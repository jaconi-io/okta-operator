/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	oktav1alpha1 "github.com/jaconi-io/okta-operator/api/v1alpha1"
	"github.com/jaconi-io/okta-operator/okta"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizerOktaClient        = "okta.jaconi.io/oktaClient"
	ConditionTypeSynced string = "Ready"
	ConditionTypeError  string = "Error"
)

// OktaClientReconciler reconciles a OktaClient object
type OktaClientReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=okta.jaconi.io,resources=oktaclients,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=okta.jaconi.io,resources=oktaclients/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=okta.jaconi.io,resources=oktaclients/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *OktaClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	oktaClient := &oktav1alpha1.OktaClient{}
	err := r.Get(ctx, req.NamespacedName, oktaClient)

	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get oktaClient %q: %w", req.NamespacedName, err)
	}

	// Handle deletion first.
	if oktaClient.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(oktaClient, finalizerOktaClient) {
			err := r.cleanUp(oktaClient, ctx, req)
			if err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(oktaClient, finalizerOktaClient)
			err = r.Update(ctx, oktaClient)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(oktaClient, finalizerOktaClient) {
		controllerutil.AddFinalizer(oktaClient, finalizerOktaClient)
		err = r.Update(ctx, oktaClient)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer %q to oktaClient %q: %w", finalizerOktaClient, req.NamespacedName, err)
		}
	}

	err = r.updateTrustedOrigins(oktaClient, ctx, req)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create or update the trusted origins %q: %w", req.NamespacedName, err)
	}

	err = r.updateApplication(oktaClient, ctx, req)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create or update application %q: %w", req.NamespacedName, err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OktaClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oktav1alpha1.OktaClient{}).
		Owns(&core.Secret{}).
		Named("oktaClient").
		Complete(r)
}

func (r *OktaClientReconciler) cleanUp(oktaClient *oktav1alpha1.OktaClient, ctx context.Context, req ctrl.Request) error {
	// Delete App
	log := ctrllog.FromContext(ctx)
	appName := oktaClient.Spec.Name
	app, err := okta.GetApplicationByLabel(appName)

	log.Info("Queried application", "appName", appName, "exists", app != nil)
	if err != nil {
		return fmt.Errorf("failed to get application %q: %w", appName, err)
	}

	if app != nil {
		log.Info("Deleting application", "appName", appName)
		err = okta.DeleteApplication(app)
		if err != nil {
			return fmt.Errorf("failed to delete application %q: %w", appName, err)
		}
	}

	// Delete trusted origins
	origins := oktaClient.Spec.TrustedOrigins
	for _, origin := range origins {
		isTrustedOrigin, err := okta.IsTrustedOrigin(origin)
		log.Info("Queried trusted origin", "origin", origin, "exists", isTrustedOrigin)
		if err != nil {
			return fmt.Errorf("failed to determine if %q is a trusted origin: %w", origin, err)
		}
		if isTrustedOrigin {
			log.Info("Deleting trusted origin", "origin", origin)
			err = okta.DeleteTrustedOrigin(origin)

			if err != nil {
				return fmt.Errorf("failed to delete trusted origin %q: %w", origin, err)
			}
		}
	}
	return nil
}

func (r *OktaClientReconciler) updateTrustedOrigins(oktaClient *oktav1alpha1.OktaClient, ctx context.Context, req ctrl.Request) error {
	// Create trusted origins
	log := ctrllog.FromContext(ctx)
	origins := oktaClient.Spec.TrustedOrigins
	for _, origin := range origins {
		isTrustedOrigin, err := okta.IsTrustedOrigin(origin)
		log.Info("Queried trusted origin", "origin", origin, "exists", isTrustedOrigin)
		if err != nil {
			return fmt.Errorf("failed to determine if %q is a trusted origin: %w", origin, err)
		}
		if isTrustedOrigin {
			// Nothing to do
			continue
		}
		log.Info("Creating trusted origin", "origin", origin)
		err = okta.CreateTrustedOrigin(origin)

		if err != nil {
			return fmt.Errorf("failed to create trusted origin %q: %w", origin, err)
		}
	}

	return nil
}

func (r *OktaClientReconciler) updateApplication(oktaClient *oktav1alpha1.OktaClient, ctx context.Context, req ctrl.Request) error {
	// Update application
	log := ctrllog.FromContext(ctx)
	secretName := oktaClient.Name
	appName := oktaClient.Spec.Name
	clientUri := oktaClient.Spec.ClientUri
	redirectUris := oktaClient.Spec.RedirectUris
	postLogoutRedirectUris := oktaClient.Spec.PostLogoutRedirectUris
	groupId := oktaClient.Spec.GroupId

	app, err := okta.GetApplicationByLabel(appName)
	log.Info("Queried application", "application", appName, "exists", app != nil)
	if err != nil {
		return fmt.Errorf("failed to get application %q: %w", appName, err)
	}

	if app == nil {
		log.Info("Creating application", "application", appName)
		app, err = okta.CreateApplication(appName, clientUri, redirectUris, postLogoutRedirectUris)
		if err != nil {
			return fmt.Errorf("failed to create application %q: %w", appName, err)
		}
	} else {
		// The application has already been created in Okta. Check if we have the client credentials for the application.
		secret := &core.Secret{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: secretName}, secret)
		if err != nil {
			if errors.IsNotFound(err) {
				// The secret does not exist, and we do not have the credentials at hand. Create a new secret.
				log.Info("Rotating application secret")
				clientSecret, err := okta.NewSecret(app.ClientID)
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
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
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
		err = okta.CreateApplicationGroupAssignment(app, groupId)
		if err != nil {
			return fmt.Errorf("failed to add application %q to group %q: %w", appName, groupId, err)
		}
	}

	return nil
}
