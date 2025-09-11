package kubeslice

import (
	"context"
	"time"

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
	startTime := time.Now()
	log.Info("ServiceImport reconcile started", 
		"namespace", req.Namespace, 
		"name", req.Name, 
		"timestamp", startTime)

	// Time the Kubernetes API call
	getStart := time.Now()
	si := &kubeslicev1beta1.ServiceImport{}
	err := r.Get(ctx, req.NamespacedName, si)
	getDuration := time.Since(getStart)
	
	if err != nil {
		log.Error(err, "Failed to get ServiceImport", 
			"namespace", req.Namespace, 
			"name", req.Name,
			"get_duration_ms", getDuration.Milliseconds())
		return reconcile.Result{}, err
	}

	log.Info("ServiceImport retrieved", 
		"namespace", si.Namespace, 
		"name", si.Name,
		"get_duration_ms", getDuration.Milliseconds(),
		"deletion_timestamp", si.ObjectMeta.DeletionTimestamp.IsZero())

	// examine DeletionTimestamp to determine if object is under deletion
	if si.ObjectMeta.DeletionTimestamp.IsZero() {
		// register our finalizer
		if !containsString(si.GetFinalizers(), finalizerName) {
			log.Info("Adding finalizer", "namespace", si.Namespace, "name", si.Name)
			finalizerStart := time.Now()
			controllerutil.AddFinalizer(si, finalizerName)
			if err := r.Update(ctx, si); err != nil {
				log.Error(err, "Failed to add finalizer", "duration_ms", time.Since(finalizerStart).Milliseconds())
				return reconcile.Result{}, err
			}
			log.Info("Finalizer added successfully", "duration_ms", time.Since(finalizerStart).Milliseconds())
			return reconcile.Result{Requeue: true}, nil
		}
	} else {
		// The object is being deleted
		if containsString(si.GetFinalizers(), finalizerName) {
			log.Info("Deleting DNS entries", "namespace", si.Namespace, "name", si.Name)
			deleteStart := time.Now()
			if err := r.EndpointsCache.Delete(si.Name, si.Spec.Slice, si.Namespace); err != nil {
				log.Error(err, "Unable to delete DNS entries", "duration_ms", time.Since(deleteStart).Milliseconds())
				return reconcile.Result{}, err
			}
			log.Info("DNS entries deleted", "duration_ms", time.Since(deleteStart).Milliseconds())

			log.Info("Removing finalizer", "namespace", si.Namespace, "name", si.Name)
			finalizerStart := time.Now()
			controllerutil.RemoveFinalizer(si, finalizerName)
			if err := r.Update(ctx, si); err != nil {
				log.Error(err, "Failed to remove finalizer", "duration_ms", time.Since(finalizerStart).Milliseconds())
				return reconcile.Result{}, err
			}
			log.Info("Finalizer removed successfully", "duration_ms", time.Since(finalizerStart).Milliseconds())
		}

		totalDuration := time.Since(startTime)
		log.Info("ServiceImport deletion reconcile completed", 
			"namespace", si.Namespace, 
			"name", si.Name,
			"total_duration_ms", totalDuration.Milliseconds())
		return reconcile.Result{}, nil
	}

	// Time the endpoint processing
	processStart := time.Now()
	eps := []slice.Endpoint{}

	log.Info("Processing ServiceImport endpoints", 
		"namespace", si.Namespace, 
		"name", si.Name,
		"endpoints_count", len(si.Status.Endpoints),
		"aliases_count", len(si.Spec.Aliases))

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
	processDuration := time.Since(processStart)

	// Time the cache update
	cacheStart := time.Now()
	r.EndpointsCache.Put(si.Name, si.Spec.Slice, si.Namespace, eps)
	cacheDuration := time.Since(cacheStart)

	// Get current cache state for logging
	allEndpoints := r.EndpointsCache.GetAll()
	
	totalDuration := time.Since(startTime)
	log.Info("ServiceImport reconcile completed", 
		"namespace", si.Namespace, 
		"name", si.Name,
		"total_duration_ms", totalDuration.Milliseconds(),
		"get_duration_ms", getDuration.Milliseconds(),
		"process_duration_ms", processDuration.Milliseconds(),
		"cache_duration_ms", cacheDuration.Milliseconds(),
		"endpoints_added", len(eps),
		"total_cached_endpoints", len(allEndpoints))

	return reconcile.Result{}, nil
}

func (r *ServiceImportReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}
