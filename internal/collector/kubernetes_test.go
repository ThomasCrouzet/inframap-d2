package collector

import (
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubernetesCollector(t *testing.T) {
	kc := &KubernetesCollector{
		TestPods:      "../../testdata/kubernetes/pods.json",
		TestServices:  "../../testdata/kubernetes/services.json",
		TestIngresses: "../../testdata/kubernetes/ingresses.json",
	}

	infra := model.NewInfrastructure()
	err := kc.Collect(infra)
	require.NoError(t, err)

	// Should have 2 namespace-servers: k8s-default and k8s-monitoring
	assert.Contains(t, infra.Servers, "k8s-default")
	assert.Contains(t, infra.Servers, "k8s-monitoring")

	// k8s-default should have nginx and postgres
	defaultServer := infra.Servers["k8s-default"]
	assert.Equal(t, model.ServerTypeCluster, defaultServer.Type)
	assert.Len(t, defaultServer.Services, 2)

	svcNames := make(map[string]bool)
	for _, svc := range defaultServer.Services {
		svcNames[svc.Name] = true
	}
	assert.True(t, svcNames["nginx"])
	assert.True(t, svcNames["postgres"])

	// Check postgres is detected as database
	for _, svc := range defaultServer.Services {
		if svc.Name == "postgres" {
			assert.Equal(t, model.ServiceTypeDatabase, svc.Type)
		}
	}

	// k8s-monitoring should have grafana
	monServer := infra.Servers["k8s-monitoring"]
	assert.Len(t, monServer.Services, 1)
	assert.Equal(t, "grafana", monServer.Services[0].Name)
}

func TestKubernetesCollectorNamespaceFilter(t *testing.T) {
	kc := &KubernetesCollector{
		Namespaces:    []string{"monitoring"},
		TestPods:      "../../testdata/kubernetes/pods.json",
		TestServices:  "../../testdata/kubernetes/services.json",
		TestIngresses: "../../testdata/kubernetes/ingresses.json",
	}

	infra := model.NewInfrastructure()
	err := kc.Collect(infra)
	require.NoError(t, err)

	// Only monitoring namespace should be present
	assert.NotContains(t, infra.Servers, "k8s-default")
	assert.Contains(t, infra.Servers, "k8s-monitoring")
}

func TestKubernetesMetadata(t *testing.T) {
	kc := &KubernetesCollector{}
	meta := kc.Metadata()
	assert.Equal(t, "kubernetes", meta.Name)
	assert.Equal(t, "kubernetes", meta.ConfigKey)
}
