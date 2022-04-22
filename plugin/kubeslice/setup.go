package kubeslice

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"os"

	meshv1beta1 "bitbucket.org/realtimeai/kubeslice-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var scheme = runtime.NewScheme()

func init() {
	clientgoscheme.AddToScheme(scheme)
	utilruntime.Must(meshv1beta1.AddToScheme(scheme))
}

// init registers this plugin.
func init() { plugin.Register("kubeslice", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	c.Next() // Ignore "example" and give us the next token.
	if c.NextArg() {
		// If there was another token, return an error, because we don't have any configuration.
		// Any errors returned from this setup function should be wrapped with plugin.Error, so we
		// can present a slightly nicer error message to the user.
		return plugin.Error("kubeslice", c.ArgErr())
	}

	ks := Kubeslice{}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		ks.Next = next
		return ks
	})

	mgr, err := manager.New(
		config.GetConfigOrDie(), manager.Options{
			Scheme: scheme,
		},
	)
	if err != nil {
		log.Error(err, "could not create manager")
		return err
	}

	err = builder.
		ControllerManagedBy(mgr).          // Create the ControllerManagedBy
		For(&meshv1beta1.ServiceImport{}). // ReplicaSet is the Application API
		Complete(&ServiceImportReconciler{
			Kubeslice: ks,
		})
	if err != nil {
		log.Error(err, "could not create controller")
		return err
	}

	// start manager in the background
	go func() {
		if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
			log.Error(err, "could not start manager")
			os.Exit(1)
		}
	}()

	// All OK, return a nil error.
	return nil
}
