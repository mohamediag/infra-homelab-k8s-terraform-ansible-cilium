package emit

import (
	"strings"
	"testing"

	"github.com/mohamediag/infra-homelab-k8s-terraform-ansible-cilium/tools/generator/internal/schema"
	"sigs.k8s.io/yaml"
)

func TestFilesBasicDev(t *testing.T) {
	svc := baseService()
	files, err := Files(svc, Options{GitOpsRepo: "/repo", ImageTag: "nginx:1.27", Envs: []string{"dev"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	assertFile(t, files, "/repo/platform-tenants-apps-dev/team-platform/sample-application/namespace.yaml", []string{
		"kind: Namespace",
		"name: sample-application-dev",
		"app.platform.homelab.io/team: team-platform",
	})
	var ns NamespaceResource
	mustUnmarshalFile(t, files, "/repo/platform-tenants-apps-dev/team-platform/sample-application/namespace.yaml", &ns)
	if ns.Metadata.Name != "sample-application-dev" {
		t.Fatalf("unexpected namespace name %q", ns.Metadata.Name)
	}

	assertFile(t, files, "/repo/platform-tenants-apps-dev/team-platform/sample-application/sample-service.yaml", []string{
		"kind: App",
		"namespace: sample-application-dev",
		"name: app-kubernetes",
		"image: nginx:1.27",
		"environment: dev",
		"LOG_LEVEL: debug",
		"cpu: 25m",
		"memory: 32Mi",
	})
	var app AppResource
	mustUnmarshalFile(t, files, "/repo/platform-tenants-apps-dev/team-platform/sample-application/sample-service.yaml", &app)
	if app.Metadata.Namespace != "sample-application-dev" {
		t.Fatalf("unexpected app namespace %q", app.Metadata.Namespace)
	}
	if app.Spec.Parameters.Config["LOG_LEVEL"] != "debug" {
		t.Fatalf("expected dev override LOG_LEVEL=debug, got %q", app.Spec.Parameters.Config["LOG_LEVEL"])
	}
}

func TestFilesSecrets(t *testing.T) {
	svc := baseService()
	svc.Secrets = []string{"DUMMY_SECRET"}
	files, err := Files(svc, Options{GitOpsRepo: "/repo", ImageTag: "nginx:1.27", Envs: []string{"dev"}})
	if err != nil {
		t.Fatal(err)
	}

	assertFile(t, files, "/repo/platform-tenants-apps-dev/team-platform/sample-application/sample-service.yaml", []string{
		"secrets:",
		"- DUMMY_SECRET",
	})
}

func TestFilesAllEnvs(t *testing.T) {
	svc := baseService()
	files, err := Files(svc, Options{GitOpsRepo: "/repo", ImageTag: "nginx:1.27"})
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 4 {
		t.Fatalf("expected 4 files, got %d", len(files))
	}
	assertFile(t, files, "/repo/platform-tenants-apps-staging/team-platform/sample-application/sample-service.yaml", []string{
		"namespace: sample-application-staging",
		"environment: staging",
		"replicas: 2",
		"LOG_LEVEL: info",
	})
}

func TestPublicExposure(t *testing.T) {
	svc := baseService()
	svc.Exposure = schema.Exposure{Type: "public", Host: "sample.example.com"}
	files, err := Files(svc, Options{GitOpsRepo: "/repo", ImageTag: "nginx:1.27", Envs: []string{"dev"}})
	if err != nil {
		t.Fatal(err)
	}

	assertFile(t, files, "/repo/platform-tenants-apps-dev/team-platform/sample-application/sample-service.yaml", []string{
		"type: public",
		"host: sample.example.com",
	})
}

func TestValidateRequiresImage(t *testing.T) {
	svc := baseService()
	svc.Image = ""
	if err := svc.Validate(""); err == nil {
		t.Fatal("expected validation error")
	}
	if err := svc.Validate("nginx:1.27"); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, files []File, path string, contains []string) {
	t.Helper()
	for _, file := range files {
		if file.Path != path {
			continue
		}
		content := string(file.Content)
		for _, want := range contains {
			if !strings.Contains(content, want) {
				t.Fatalf("%s does not contain %q:\n%s", path, want, content)
			}
		}
		return
	}
	t.Fatalf("file %s not found", path)
}

func mustUnmarshalFile(t *testing.T, files []File, path string, out any) {
	t.Helper()
	for _, file := range files {
		if file.Path != path {
			continue
		}
		if err := yaml.UnmarshalStrict(file.Content, out); err != nil {
			t.Fatalf("unmarshal %s: %v", path, err)
		}
		return
	}
	t.Fatalf("file %s not found", path)
}

func baseService() *schema.Service {
	return &schema.Service{
		Name:        "sample-service",
		Application: "sample-application",
		Team:        "team-platform",
		Port:        80,
		HealthCheck: schema.HealthCheck{Liveness: "/", Readiness: "/"},
		Metrics:     schema.Metrics{Enabled: false, Path: "/metrics"},
		Exposure:    schema.Exposure{Type: "none", Host: "sample-service.example.com"},
		Config: schema.Config{
			Default:   map[string]string{"LOG_LEVEL": "info"},
			Overrides: map[string]map[string]string{"dev": {"LOG_LEVEL": "debug"}},
		},
		Environments: map[string]schema.Environment{
			"dev": {
				Replicas: 1,
				Resources: schema.Resources{
					CPU:    "25m/100m",
					Memory: "32Mi/64Mi",
				},
			},
			"staging": {
				Replicas: 2,
				Resources: schema.Resources{
					CPU:    "50m/200m",
					Memory: "64Mi/128Mi",
				},
			},
		},
	}
}
