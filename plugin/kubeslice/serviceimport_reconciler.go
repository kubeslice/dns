package kubeslice

import (
	"context"
	"strings"

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

func getSrvEntries(si *kubeslicev1beta1.ServiceImport, name, ip string) []slice.Endpoint {
	srvEntries := []slice.Endpoint{}
	for _, port := range si.Spec.Ports {
		srvEntry := slice.Endpoint{
			Host:        "_" + port.Name + "." + "_" + strings.ToLower(string(port.Protocol)) + "." + name,
			TargetStrip: 2,
			IP:          ip,
			Ports: []slice.ServicePort{{
				Name:     port.Name,
				Port:     port.ContainerPort,
				Protocol: string(port.Protocol),
			}},
		}
		srvEntries = append(srvEntries, srvEntry)
	}

	return srvEntries
}

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
	ports := []slice.ServicePort{}
	for _, port := range si.Spec.Ports {
		ports = append(ports, slice.ServicePort{
			Name:     port.Name,
			Port:     port.ContainerPort,
			Protocol: string(port.Protocol),
		})
	}

	for _, ep := range si.Status.Endpoints {
		// Add entries for the service dns name
		endpoint := slice.Endpoint{
			Host:  ep.DNSName,
			IP:    ep.IP,
			Ports: ports,
		}
		endpoint2 := slice.Endpoint{
			Host:  si.Spec.DNSName,
			IP:    ep.IP,
			Ports: ports,
		}
		eps = append(eps, endpoint, endpoint2)
		// Add port number entries for SRV records
		eps = append(eps, getSrvEntries(si, si.Spec.DNSName, ep.IP)...)

		// Add entries for aliases for the service
		for _, alias := range si.Spec.Aliases {
			endpointN := slice.Endpoint{
				Host:  alias,
				IP:    ep.IP,
				Ports: ports,
			}
			eps = append(eps, endpointN)
			// Add port number entries for SRV records
			eps = append(eps, getSrvEntries(si, alias, ep.IP)...)
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
