# Kubeslice DNS design and architecture

Kubeslice DNS is implemented as a coredns plugin to facilitate service
discovery in kubeslice

## Service Discovery

When a service is exported through `ServiceExport` CRD in a cluster, the
endpoints corresponding to the service are populated in the CR by the
operator. This ServiceExport is then pushed to the hub cluster and then
hub cluster transforms it into a SpokeServiceImport CR which can be
imported in all the clusters. This CR contains endpoints corresponding
to a particular service across all available clusters.

This SpokeServiceImport is our source of truth for populating DNS
entries. Kubeslice operator creates a ServiceImport resource for this CR
received from hub.

Here is a sample ServiceImport CR

```
apiVersion: mesh.avesha.io/v1beta1
kind: ServiceImport
metadata:
  name: nginx
  namespace: default
spec:
  slice: green
  dnsName: nginx.default.svc.slice.local
  ports:
  - name: http
    containerPort: 80
    protocol: TCP
status:
  importStatus: READY
  endpoints:
  - name: nginx-d955f6db-9nlxm
    ip: 10.7.1.94
    port: 80
    clusterId: jd-cluster-7
    dnsName: nginx-d955f6db-9nlxm.cluster-7-jd.nginx.default.svc.slice.local
  - name: nginx-d955f6db-wjrjv
    ip: 10.7.2.59
    port: 80
    clusterId: jd-cluster-7
    dnsName: nginx-d955f6db-wjrjv.cluster-7-jd.nginx.default.svc.slice.local
```

There is a dnsName corresponding to the service. When queried for this
url, dns server should return IP address corresponding to all the
endpoints.

Additionally, each endpoints has its own dns name. DNS server should
answer these queries as well.

## CoreDNS plugin design

References:

Example plugin:
https://github.com/coredns/example

Plugin writing guide:
https://coredns.io/2016/12/19/writing-plugins-for-coredns/

External plugins manual
https://coredns.io/manual/explugins/

Kubernetes Plugin
https://coredns.io/plugins/kubernetes/

Usually, to develop a coredns plugin we have to clone the coredns repo,
add plugin code to a subfolder and declare the plugin in a config file.
But it is hard to maintain. So in this case we have created an
independent project, imported the coredns as a dependency and hooked our
code into it as a plugin. Then coredns is executed with kubeslice plugin


The plugin code is in `plugin/kubeslice` directory. It can be enabled by
adding `kubeslice` as a plugin in corefile. 

Example:

```
.:1053 {
  forward . 9.9.9.9
  kubeslice
}
```

In order to register a struct as coredns plugin, it should implement the
plugin.Handler interface. The `ready` method should return `true` when
it is ready to handle requests. In our case, we always return true.
Ideally we could wait for initial sync to complete before setting it to
true. `ServiceDNS` method in the Handler interface is responsible for
anwering dns queries and returning responses.

Additionally, the package should have a `setup` function. We initialize
the plugin, start our controllers and add the plugin to coredns plugin
chain to be able to handle requests.

Currently, kubeslice plugin supports only `A` records (https://support.dnsimple.com/articles/a-record/).
We can think about adding reverse DNS lookups and srv records in the
future.

## Endpoints Cache

To avoid querying the kubernetes api server for each dns query, an
endpoints cache was introduced. It is implemented as a simple map to store list of
endpoints corresponding to each host

```
type EndpointsCache interface {
	GetAll() []slice.Endpoint
	Get(name, slice, namespace string) []slice.Endpoint
	Put(name, slice, namespace string, endpints []slice.Endpoint) error
	Delete(name, slice, namespace string) error
}

// Implement EndpointsCache
type endpointsCache struct {
	cache map[string][]slice.Endpoint
}

// Endpoint corresponds to a dns entry for an endpint in the slice
type Endpoint struct {
	Host string
	IP   string
}
```

## Kubernetes controller

A kubernetes controller was implemented using `controller-runtime`. It
watches for `ServiceImport` entries, and updates the entries in cache
whenever there is a change in ServiceImport or a new ServiceImport is
added.

The controller is setup in the setup method of the plugin and runs in a
separate thread. ServceDNS method uses cache as a source of truth for
handling requests.

## Working of the plugin

Here is the high level overview of how the plugin works:

1. Setup controller manager with access to local cluster
1. Start the controller
1. Add the plugin to coredns handler chain
1. Controller watches for ServiceImport resources
1. When a new ServiceImport resource is found, add the list of endpoints
   to cache
1. When ServiceImport resource is updated, update the same in cache
1. When a dns requst arrives, check the cache and return response

Even though the cache implementation is a just a map, it should be able
to handle thousands of services.
