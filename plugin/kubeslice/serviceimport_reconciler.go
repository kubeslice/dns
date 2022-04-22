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
	Kubeslice Kubeslice
}

// Watch the ServiceImport changes and adjust dns cache accordingly
func (r *ServiceImportReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	si := &meshv1beta1.ServiceImport{}
	err := r.Get(ctx, req.NamespacedName, si)
	if err != nil {
		return reconcile.Result{}, err
	}

	log.Info("got si")

	for _, ep := range si.Status.Endpoints {
		endpoint := SliceEndpoint{
			Host: ep.DNSName,
			IP:   ep.IP,
		}
		r.Kubeslice.SliceEndpoints = append(r.Kubeslice.SliceEndpoints, endpoint)
	}

	log.Info(r.Kubeslice.SliceEndpoints)

	return reconcile.Result{}, nil
}

func (r *ServiceImportReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}
