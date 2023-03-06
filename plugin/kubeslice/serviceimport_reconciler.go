package kubeslice

import (
	"context"

	dnsCache "github.com/kubeslice/dns/plugin/kubeslice/cache"
	"github.com/kubeslice/dns/plugin/kubeslice/slice"
	kubeslicev1beta1 "github.com/kubeslice/worker-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReplicaSetReconciler is a simple ControllerManagedBy example implementation.
type ServiceImportReconciler struct {
	client.Client
	EndpointsCache dnsCache.EndpointsCache
}

const finalizerName = "networking.kubeslice.io/dns-finalizer"

// Watch the ServiceImport changes and adjust dns cache accordingly
func (r *ServiceImportReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	si := &kubeslicev1beta1.ServiceImport{}
	err := r.Get(ctx, req.NamespacedName, si)
	if err != nil {
		return reconcile.Result{}, err
	}

	log.Info("got si")

	// examine DeletionTimestamp to determine if object is under deletion
	if si.ObjectMeta.DeletionTimestamp.IsZero() {
		// register our finalizer
		if !containsString(si.GetFinalizers(), finalizerName) {
			log.Info("adding finalizer")
			controllerutil.AddFinalizer(si, finalizerName)
			if err := r.Update(ctx, si); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	} else {
		// The object is being deleted
		if containsString(si.GetFinalizers(), finalizerName) {
			log.Info("deleting dns entries")
			if err := r.EndpointsCache.Delete(si.Name, si.Spec.Slice, si.Namespace); err != nil {
				log.Error(err, "unable to delete dns entries")
				return reconcile.Result{}, err
			}

			log.Info("removing finalizer")
			controllerutil.RemoveFinalizer(si, finalizerName)
			if err := r.Update(ctx, si); err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, nil
	}

	eps := []slice.Endpoint{}

	for _, ep := range si.Status.Endpoints {
		endpoint := slice.Endpoint{
			Host: ep.DNSName,
			IP:   ep.IP,
		}
		endpoint2 := slice.Endpoint{
			Host: si.Spec.DNSName,
			IP:   ep.IP,
		}
		eps = append(eps, endpoint, endpoint2)
		for _, alias := range si.Spec.Aliases {
                        endpointN := slice.Endpoint{
			       Host: alias,
			       IP:   ep.IP,
		        }
		        eps = append(eps, endpointN)
		}
	}

	r.EndpointsCache.Put(si.Name, si.Spec.Slice, si.Namespace, eps)

	log.Info(r.EndpointsCache.GetAll())

	return reconcile.Result{}, nil
}

func (r *ServiceImportReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}
