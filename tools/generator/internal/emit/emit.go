package emit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mohamediag/infra-homelab-k8s-terraform-ansible-cilium/tools/generator/internal/schema"
	"sigs.k8s.io/yaml"
)

const (
	appAPIVersion = "platform.homelab.io/v1alpha1"
	composition   = "app-kubernetes"
)

type File struct {
	Path    string
	Content []byte
}

type Options struct {
	GitOpsRepo string
	ImageTag   string
	Envs       []string
}

func Write(files []File) error {
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
			return fmt.Errorf("create parent directory for %s: %w", f.Path, err)
		}
		if err := os.WriteFile(f.Path, f.Content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", f.Path, err)
		}
	}
	return nil
}

func RelativePath(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return rel
}

func Files(svc *schema.Service, opts Options) ([]File, error) {
	envs := opts.Envs
	if len(envs) == 0 {
		envs = schema.EnvOrder(svc.Environments)
	}

	files := make([]File, 0, len(envs)*2)
	for _, env := range envs {
		cfg, ok := svc.Environments[env]
		if !ok {
			return nil, fmt.Errorf("environment %q not defined in service contract", env)
		}

		ns, err := marshal(Namespace(svc, env))
		if err != nil {
			return nil, err
		}
		appResource, err := App(svc, env, cfg, opts.ImageTag)
		if err != nil {
			return nil, err
		}
		app, err := marshal(appResource)
		if err != nil {
			return nil, err
		}

		base := filepath.Join(opts.GitOpsRepo, fmt.Sprintf("platform-tenants-apps-%s", env), svc.Team, svc.Application)
		files = append(files, File{
			Path:    filepath.Join(base, "namespace.yaml"),
			Content: ns,
		})
		files = append(files, File{
			Path:    filepath.Join(base, svc.Name+".yaml"),
			Content: app,
		})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func NamespaceName(svc *schema.Service, env string) string {
	return fmt.Sprintf("%s-%s", svc.Application, env)
}

func marshal(v any) ([]byte, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	b = bytes.ReplaceAll(b, []byte("\n{}\n"), []byte("\n"))
	return append(b, '\n'), nil
}

type Metadata struct {
	Name        string            `json:"name" yaml:"name"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type NamespaceResource struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
}

func Namespace(svc *schema.Service, env string) NamespaceResource {
	return NamespaceResource{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata: Metadata{
			Name: NamespaceName(svc, env),
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": "-10",
			},
			Labels: map[string]string{
				"app.platform.homelab.io/team":        svc.Team,
				"app.platform.homelab.io/application": svc.Application,
				"app.platform.homelab.io/environment": env,
			},
		},
	}
}

type AppResource struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	Spec       AppSpec  `json:"spec" yaml:"spec"`
}

type AppSpec struct {
	Crossplane AppCrossplane `json:"crossplane" yaml:"crossplane"`
	Parameters AppParameters `json:"parameters" yaml:"parameters"`
}

type AppCrossplane struct {
	CompositionRef NameRef `json:"compositionRef" yaml:"compositionRef"`
}

type NameRef struct {
	Name string `json:"name" yaml:"name"`
}

type AppParameters struct {
	Name        string             `json:"name" yaml:"name"`
	Owner       string             `json:"owner" yaml:"owner"`
	Environment string             `json:"environment" yaml:"environment"`
	Image       string             `json:"image" yaml:"image"`
	Port        int                `json:"port" yaml:"port"`
	Replicas    int                `json:"replicas" yaml:"replicas"`
	Resources   AppResources       `json:"resources" yaml:"resources"`
	Config      map[string]string  `json:"config" yaml:"config"`
	Secrets     []string           `json:"secrets" yaml:"secrets"`
	Exposure    schema.Exposure    `json:"exposure" yaml:"exposure"`
	HealthCheck schema.HealthCheck `json:"healthCheck" yaml:"healthCheck"`
	Metrics     schema.Metrics     `json:"metrics" yaml:"metrics"`
}

type AppResources struct {
	Requests ResourceValues `json:"requests" yaml:"requests"`
	Limits   ResourceValues `json:"limits" yaml:"limits"`
}

type ResourceValues struct {
	CPU    string `json:"cpu" yaml:"cpu"`
	Memory string `json:"memory" yaml:"memory"`
}

func App(svc *schema.Service, env string, cfg schema.Environment, imageTag string) (AppResource, error) {
	cpu, err := schema.ParseResourcePair(cfg.Resources.CPU)
	if err != nil {
		return AppResource{}, fmt.Errorf("parse cpu resources for %s: %w", env, err)
	}
	memory, err := schema.ParseResourcePair(cfg.Resources.Memory)
	if err != nil {
		return AppResource{}, fmt.Errorf("parse memory resources for %s: %w", env, err)
	}

	return AppResource{
		APIVersion: appAPIVersion,
		Kind:       "App",
		Metadata: Metadata{
			Name:      svc.Name,
			Namespace: NamespaceName(svc, env),
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": "0",
			},
		},
		Spec: AppSpec{
			Crossplane: AppCrossplane{CompositionRef: NameRef{Name: composition}},
			Parameters: AppParameters{
				Name:        svc.Name,
				Owner:       svc.Team,
				Environment: env,
				Image:       svc.ImageFor(imageTag),
				Port:        svc.Port,
				Replicas:    cfg.Replicas,
				Resources: AppResources{
					Requests: ResourceValues{CPU: cpu.Request, Memory: memory.Request},
					Limits:   ResourceValues{CPU: cpu.Limit, Memory: memory.Limit},
				},
				Config:  svc.ConfigFor(env),
				Secrets: svc.Secrets,
				Exposure: schema.Exposure{
					Type: svc.ExposureType(),
					Host: svc.Exposure.Host,
				},
				HealthCheck: svc.HealthCheck,
				Metrics: schema.Metrics{
					Enabled: svc.Metrics.Enabled,
					Path:    svc.MetricsPath(),
				},
			},
		},
	}, nil
}
