package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"github.com/jaconi-io/okta-operator/okta"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	finalizerTrustedOrigin  = "okta.jaconi.io/trusted-origin"
	annotationTrustedOrigin = "okta.jaconi.io/trusted-origin"
)

// TrustedOriginReconciler manages Okta trusted origins for annotated ingress resources.
type TrustedOriginReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (r *TrustedOriginReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	ingress := &networkingv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get ingress %q: %w", req.NamespacedName, err)
	}

	// Determine the trusted origin based on the okta.jaconi.io/trusted-origin annotation.
	origin, ok := ingress.Annotations[annotationTrustedOrigin]
	if !ok {
		// This should never happen! Check the controllers filter logic if you ever encounter this error.
		return ctrl.Result{}, fmt.Errorf("missing annotation %q for ingress %q", annotationTrustedOrigin, req.NamespacedName)
	}

	// Sleep to make sure Okta replication has happened
	time.Sleep(2 * time.Second)

	isTrustedOrigin, err := okta.IsTrustedOrigin(origin)
	log.Info("Queried trusted origin", "isTrustedOrigin", isTrustedOrigin)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to determine if %q (from ingress %q) is a trusted origin: %w", origin, req.NamespacedName, err)
	}

	// Handle deletion first.
	if ingress.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(ingress, finalizerTrustedOrigin) {
			if isTrustedOrigin {
				log.Info("Deleting trusted origin")
				err = okta.DeleteTrustedOrigin(origin)

				if err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to delete trusted origin %q (from ingress %q): %w", origin, req.NamespacedName, err)
				}
			}

			controllerutil.RemoveFinalizer(ingress, finalizerTrustedOrigin)
			err := r.Update(ctx, ingress)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(ingress, finalizerTrustedOrigin) {
		controllerutil.AddFinalizer(ingress, finalizerTrustedOrigin)
		err = r.Update(ctx, ingress)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer %q to ingress %q: %w", finalizerTrustedOrigin, req.NamespacedName, err)
		}
	}

	// Nothing to do if origin is already trusted.
	if isTrustedOrigin {
		return ctrl.Result{}, nil
	}

	log.Info("Creating trusted origin")
	err = okta.CreateTrustedOrigin(origin)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create trusted origin %q for ingress %q: %w", origin, req.NamespacedName, err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrustedOriginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	hasAnnotation := func(o client.Object) bool {
		_, ok := o.GetAnnotations()[annotationTrustedOrigin]
		return ok
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Named("trustedOrigin").
		WithEventFilter(predicate.NewPredicateFuncs(hasAnnotation)).
		Complete(r)
}
