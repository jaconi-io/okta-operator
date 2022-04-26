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
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const ingressFinalizer = "okta.jaconi.io/finalizer"
const oktaAnno = "okta.jaconi.io/active"

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	log.Info("Ingress Event occured!")

	// Fetch the Ingress instance
	ingress := &networkingv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ingress)

	if err != nil {
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Ingress")
		return ctrl.Result{}, err
	}

	// TODO: Check if the Okta entry already exists, if not create a new one
	hosts := extractHosts(ingress)
	fmt.Println(hosts)

	// Check if the Ingress instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isIngressMarkedToBeDeleted := ingress.GetDeletionTimestamp() != nil
	if isIngressMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(ingress, ingressFinalizer) {
			// Run finalization logic for memcachedFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeIngress(log, ingress); err != nil {
				return ctrl.Result{}, err
			}

			// Remove ingressFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(ingress, ingressFinalizer)
			err := r.Update(ctx, ingress)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this Ingress
	if !controllerutil.ContainsFinalizer(ingress, ingressFinalizer) {
		controllerutil.AddFinalizer(ingress, ingressFinalizer)
		err = r.Update(ctx, ingress)
		log.Info("Finalizer added to Ingress")
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) finalizeIngress(reqLogger logr.Logger, m *networkingv1.Ingress) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	reqLogger.Info("Successfully finalized Ingress", "host", m.Spec.Rules[0].Host)
	return nil
}

func (r *IngressReconciler) containsAnnotation(o client.Object) bool {
	return o.GetAnnotations()[oktaAnno] == "true"
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				// The reconciler adds a finalizer so we perform clean-up
				// when the delete timestamp is added
				// Suppress Delete events to avoid filtering them out in the Reconcile function
				fmt.Println("Delete Event")
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				fmt.Println("Update Event")
				return r.containsAnnotation(e.ObjectNew)
			},
			CreateFunc: func(e event.CreateEvent) bool {
				fmt.Println("Create Event")
				return r.containsAnnotation(e.Object)
			},
		}).
		Complete(r)
}
