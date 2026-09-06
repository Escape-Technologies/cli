package kube

import "testing"

func TestIsInCluster(t *testing.T) {
	t.Run("without Kubernetes service environment variables", func(t *testing.T) {
		t.Setenv("KUBERNETES_SERVICE_HOST", "")
		t.Setenv("KUBERNETES_SERVICE_PORT", "")

		if isInCluster() {
			t.Fatal("expected false")
		}
	})

	t.Run("with Kubernetes service environment variables", func(t *testing.T) {
		t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
		t.Setenv("KUBERNETES_SERVICE_PORT", "443")

		if !isInCluster() {
			t.Fatal("expected true")
		}
	})
}
