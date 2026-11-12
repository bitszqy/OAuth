package lab00

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cyllective/oauth-labs/configure/internal/constants"
)

func TestNumberedConfig(t *testing.T) {
	client := numberedConfig(clientConfigTemplate, "01")
	server := numberedConfig(serverConfigTemplate, "01")
	for _, expected := range []string{"client01", "client-01.oauth.labs", "server-01.oauth.labs"} {
		if !strings.Contains(client, expected) {
			t.Errorf("client config missing %q", expected)
		}
	}
	for _, expected := range []string{"server01", "server-01.oauth.labs"} {
		if !strings.Contains(server, expected) {
			t.Errorf("server config missing %q", expected)
		}
	}
	if strings.Contains(client, "client00") || strings.Contains(server, "server00") {
		t.Fatal("baseline database name leaked into new lab")
	}
	if !strings.Contains(client, "max_age: 80400") {
		t.Fatal("unrelated numeric configuration was changed")
	}
}

func TestLab01ClientConfigIncludesCalendarScope(t *testing.T) {
	original := constants.DockerDir
	defer func() { constants.DockerDir = original }()
	constants.DockerDir = t.TempDir()

	Configure("01")

	config, err := os.ReadFile(filepath.Join(constants.DockerDir, "lab01", "client.config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(config), "read:calendar") {
		t.Error("lab01 client config is missing the calendar scope")
	}
}
