package controllers

import (
	"context"
	"fmt"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"github.com/jaconi-io/okta-operator/okta"

	core "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	finalizerApplication  = "okta.jaconi.io/application"
	annotationApplication = "okta.jaconi.io/application"

	oktaClientSecretName = "okta-client"
)

// ApplicationReconciler manages Okta applications for annotated ingress resources.
type ApplicationReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	GroupID string
}

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	ingress := &networkingv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get ingress %q: %w", req.NamespacedName, err)
	}

	// Determine the application based on the okta.jaconi.io/application annotation.
	application, ok := ingress.Annotations[annotationApplication]
	if !ok {
		// This should never happen! Check the controllers filter logic if you ever encounter this error.
		return ctrl.Result{}, fmt.Errorf("missing annotation %q for ingress %q", annotationApplication, req.NamespacedName)
	}

	// Sleep to make sure Okta replication has happened
	time.Sleep(2 * time.Second)

	app, err := okta.GetApplicationByLabel(application)
	log.Info("Queried application", "applicationExists", app != nil)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get application %q: %w", application, err)
	}

	// Handle deletion first.
	if ingress.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(ingress, finalizerApplication) {
			if app != nil {
				log.Info("Deleting application")
				err = okta.DeleteApplication(app)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to delete application %q: %w", application, err)
				}
			}

			controllerutil.RemoveFinalizer(ingress, finalizerApplication)
			err := r.Update(ctx, ingress)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(ingress, finalizerApplication) {
		controllerutil.AddFinalizer(ingress, finalizerApplication)
		err = r.Update(ctx, ingress)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer %q to ingress %q: %w", finalizerApplication, req.NamespacedName, err)
		}
	}

	if app == nil {
		log.Info("Creating application")
		app, err = okta.CreateApplication(application)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create application %q for ingress %q: %w", application, req.NamespacedName, err)
		}
	} else {
		// The application has already been created in Okta. Check if we have the client credentials for the application.
		secret := &core.Secret{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: oktaClientSecretName}, secret)
		if err != nil {
			if errors.IsNotFound(err) {
				// The secret does not exist, and we do not have the credentials at hand. Create a new secret.
				log.Info("Rotating application secret")
				clientSecret, err := okta.NewSecret(app.ClientID)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("could not rotate secret for application %q: %w", application, err)
				}

				app.ClientSecret = clientSecret
			}
		}
	}

	if r.GroupID != "" {
		err = okta.CreateApplicationGroupAssignment(app, r.GroupID)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add application %q to group %q: %w", application, r.GroupID, err)
		}
	}

	// If we have a new ClientSecret, create or update the K8s secret
	if app.ClientSecret != "" {
		secret := &core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      oktaClientSecretName,
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
		return ctrl.Result{}, fmt.Errorf("failed to create / update secret for application %q: %w", application, err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	hasAnnotation := func(o client.Object) bool {
		_, ok := o.GetAnnotations()[annotationApplication]
		return ok
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Named("application").
		Owns(&core.Secret{}).
		WithEventFilter(predicate.NewPredicateFuncs(hasAnnotation)).
		Complete(r)
}
