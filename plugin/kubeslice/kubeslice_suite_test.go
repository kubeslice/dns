package kubeslice_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubeslice(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubeslice Suite")
}
