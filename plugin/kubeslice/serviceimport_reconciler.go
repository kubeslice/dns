package kubeslice

import (
	"context"

	meshv1beta1 "bitbucket.org/realtimeai/kubeslice-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReplicaSetReconciler is a simple ControllerManagedBy example implementation.
type ServiceImportReconciler struct {
	client.Client
}

// Implement the business logic:
// This function will be called when there is a change to a ReplicaSet or a Pod with an OwnerReference
// to a ReplicaSet.
//
// * Read the ReplicaSet
// * Read the Pods
// * Set a Label on the ReplicaSet with the Pod count.
func (r *ServiceImportReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	si := &meshv1beta1.ServiceImport{}
	err := r.Get(ctx, req.NamespacedName, si)
	if err != nil {
		return reconcile.Result{}, err
	}

	log.Info("got si")

	return reconcile.Result{}, nil
}

func (r *ServiceImportReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}
