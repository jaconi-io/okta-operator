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
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	err := r.deleteApplication(oktaClient, ctx)
	if err != nil {
		return err
	}

	// Delete trusted origins
	err = r.deleteTrustedOrigins(oktaClient, ctx)
	if err != nil {
		return err
	}
	return nil
}
