package collector

import (
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemdCollector(t *testing.T) {
	sc := &SystemdCollector{
		Servers: []systemdServer{
			{
				Host:     "myserver",
				TestFile: "../../testdata/systemd/units.json",
			},
		},
	}

	infra := model.NewInfrastructure()
	err := sc.Collect(infra)
	require.NoError(t, err)

	assert.Contains(t, infra.Servers, "myserver")
	server := infra.Servers["myserver"]

	// Should have all 6 services from the test data
	assert.Len(t, server.Services, 6)

	svcNames := make(map[string]bool)
	for _, svc := range server.Services {
		svcNames[svc.Name] = true
	}
	assert.True(t, svcNames["docker"])
	assert.True(t, svcNames["nginx"])
	assert.True(t, svcNames["sshd"])
	assert.True(t, svcNames["postgresql"])
}

func TestSystemdCollectorWithFilter(t *testing.T) {
	sc := &SystemdCollector{
		Servers: []systemdServer{
			{
				Host:     "filtered",
				Filter:   []string{"docker", "nginx"},
				TestFile: "../../testdata/systemd/units.json",
			},
		},
	}

	infra := model.NewInfrastructure()
	err := sc.Collect(infra)
	require.NoError(t, err)

	server := infra.Servers["filtered"]
	assert.Len(t, server.Services, 2)
}

func TestSystemdCollectorWithExclude(t *testing.T) {
	sc := &SystemdCollector{
		Servers: []systemdServer{
			{
				Host:     "excluded",
				Exclude:  []string{"cron", "network", "sshd"},
				TestFile: "../../testdata/systemd/units.json",
			},
		},
	}

	infra := model.NewInfrastructure()
	err := sc.Collect(infra)
	require.NoError(t, err)

	server := infra.Servers["excluded"]
	// 6 total - 3 excluded = 3 remaining
	assert.Len(t, server.Services, 3)

	svcNames := make(map[string]bool)
	for _, svc := range server.Services {
		svcNames[svc.Name] = true
	}
	assert.False(t, svcNames["cron"])
	assert.False(t, svcNames["networkd"])
	assert.False(t, svcNames["sshd"])
}

func TestSystemdCollectorPostgresType(t *testing.T) {
	sc := &SystemdCollector{
		Servers: []systemdServer{
			{
				Host:     "dbserver",
				Filter:   []string{"postgresql"},
				TestFile: "../../testdata/systemd/units.json",
			},
		},
	}

	infra := model.NewInfrastructure()
	err := sc.Collect(infra)
	require.NoError(t, err)

	server := infra.Servers["dbserver"]
	require.Len(t, server.Services, 1)
	assert.Equal(t, model.ServiceTypeDatabase, server.Services[0].Type)
}

func TestSystemdMetadata(t *testing.T) {
	sc := &SystemdCollector{}
	meta := sc.Metadata()
	assert.Equal(t, "systemd", meta.Name)
	assert.Equal(t, "systemd", meta.ConfigKey)
}
