package schema

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"
)

var namePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

var allowedEnvs = []string{"dev", "staging", "prod"}

type Service struct {
	Name         string                 `json:"name" yaml:"name"`
	Application  string                 `json:"application" yaml:"application"`
	Team         string                 `json:"team" yaml:"team"`
	Description  string                 `json:"description" yaml:"description"`
	CI           *CI                    `json:"ci" yaml:"ci"`
	Image        string                 `json:"image" yaml:"image"`
	Port         int                    `json:"port" yaml:"port"`
	HealthCheck  HealthCheck            `json:"healthCheck" yaml:"healthCheck"`
	Metrics      Metrics                `json:"metrics" yaml:"metrics"`
	Exposure     Exposure               `json:"exposure" yaml:"exposure"`
	Config       Config                 `json:"config" yaml:"config"`
	Secrets      []string               `json:"secrets" yaml:"secrets"`
	Environments map[string]Environment `json:"environments" yaml:"environments"`
}

type CI struct {
	AutoCommitDev bool `json:"autoCommitDev" yaml:"autoCommitDev"`
}

type HealthCheck struct {
	Liveness  string `json:"liveness" yaml:"liveness"`
	Readiness string `json:"readiness" yaml:"readiness"`
}

type Metrics struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Path    string `json:"path" yaml:"path"`
}

type Exposure struct {
	Type string `json:"type" yaml:"type"`
	Host string `json:"host" yaml:"host"`
}

type Config struct {
	Default   map[string]string            `json:"default" yaml:"default"`
	Overrides map[string]map[string]string `json:"overrides" yaml:"overrides"`
}

type Environment struct {
	Replicas  int       `json:"replicas" yaml:"replicas"`
	Resources Resources `json:"resources" yaml:"resources"`
}

type Resources struct {
	CPU    string `json:"cpu" yaml:"cpu"`
	Memory string `json:"memory" yaml:"memory"`
}

type ResourcePair struct {
	Request string
	Limit   string
}

func Load(path string) (*Service, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var svc Service
	if err := yaml.UnmarshalStrict(b, &svc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &svc, nil
}

func (s *Service) Validate(imageTag string) error {
	var errs []string

	if s.Name == "" {
		errs = append(errs, "name is required")
	} else if !namePattern.MatchString(s.Name) {
		errs = append(errs, "name must be a DNS label")
	}

	if s.Application == "" {
		errs = append(errs, "application is required")
	} else if !namePattern.MatchString(s.Application) {
		errs = append(errs, "application must be a DNS label")
	}

	if s.Team == "" {
		errs = append(errs, "team is required")
	} else if !namePattern.MatchString(s.Team) {
		errs = append(errs, "team must be a DNS label")
	}

	if s.Port < 1 || s.Port > 65535 {
		errs = append(errs, "port must be between 1 and 65535")
	}

	if s.Image == "" && imageTag == "" {
		errs = append(errs, "--image-tag or image in service.yaml is required")
	}

	if s.HealthCheck.Liveness == "" {
		errs = append(errs, "healthCheck.liveness is required")
	}
	if s.HealthCheck.Readiness == "" {
		errs = append(errs, "healthCheck.readiness is required")
	}
	if s.Metrics.Enabled && s.Metrics.Path == "" {
		errs = append(errs, "metrics.path is required when metrics.enabled=true")
	}

	switch s.Exposure.Type {
	case "", "none", "private", "public":
	case "None", "Private", "Public":
		errs = append(errs, "exposure.type must be lowercase: none, private, or public")
	default:
		errs = append(errs, "exposure.type must be one of: none, private, public")
	}
	if s.Exposure.Type == "public" && s.Exposure.Host == "" {
		errs = append(errs, "exposure.host is required when exposure.type=public")
	}

	if len(s.Environments) == 0 {
		errs = append(errs, "environments must contain at least one environment")
	}
	for env, cfg := range s.Environments {
		if !IsAllowedEnv(env) {
			errs = append(errs, fmt.Sprintf("environment %q must be one of: dev, staging, prod", env))
		}
		if cfg.Replicas < 1 {
			errs = append(errs, fmt.Sprintf("environments.%s.replicas must be >= 1", env))
		}
		if _, err := ParseResourcePair(cfg.Resources.CPU); err != nil {
			errs = append(errs, fmt.Sprintf("environments.%s.resources.cpu: %v", env, err))
		}
		if _, err := ParseResourcePair(cfg.Resources.Memory); err != nil {
			errs = append(errs, fmt.Sprintf("environments.%s.resources.memory: %v", env, err))
		}
	}

	for env := range s.Config.Overrides {
		if !IsAllowedEnv(env) {
			errs = append(errs, fmt.Sprintf("config.overrides.%s must be one of: dev, staging, prod", env))
		}
	}

	if len(errs) > 0 {
		sort.Strings(errs)
		return fmt.Errorf("invalid service contract:\n- %s", strings.Join(errs, "\n- "))
	}

	return nil
}

func (s *Service) ImageFor(imageTag string) string {
	if imageTag != "" {
		return imageTag
	}
	return s.Image
}

func (s *Service) ExposureType() string {
	if s.Exposure.Type == "" {
		return "none"
	}
	return s.Exposure.Type
}

func (s *Service) MetricsPath() string {
	if s.Metrics.Path == "" {
		return "/metrics"
	}
	return s.Metrics.Path
}

func (s *Service) ConfigFor(env string) map[string]string {
	out := make(map[string]string, len(s.Config.Default)+len(s.Config.Overrides[env]))
	for k, v := range s.Config.Default {
		out[k] = v
	}
	for k, v := range s.Config.Overrides[env] {
		out[k] = v
	}
	return out
}

func (s *Service) AutoCommitDev() bool {
	if s.CI == nil {
		return true
	}
	return s.CI.AutoCommitDev
}

func IsAllowedEnv(env string) bool {
	for _, allowed := range allowedEnvs {
		if env == allowed {
			return true
		}
	}
	return false
}

func EnvOrder(envs map[string]Environment) []string {
	out := make([]string, 0, len(envs))
	for _, env := range allowedEnvs {
		if _, ok := envs[env]; ok {
			out = append(out, env)
		}
	}
	return out
}

func SelectEnvs(serviceEnvs map[string]Environment, env string, all bool) ([]string, error) {
	if env != "" && all {
		return nil, fmt.Errorf("use either --env or --all-envs, not both")
	}
	if env == "" && !all {
		return nil, fmt.Errorf("one of --env or --all-envs is required")
	}
	if env != "" {
		if !IsAllowedEnv(env) {
			return nil, fmt.Errorf("--env must be one of: dev, staging, prod")
		}
		if _, ok := serviceEnvs[env]; !ok {
			return nil, fmt.Errorf("environment %q not defined in service contract", env)
		}
		return []string{env}, nil
	}
	return EnvOrder(serviceEnvs), nil
}

func ParseResourcePair(value string) (ResourcePair, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return ResourcePair{}, fmt.Errorf("must use request/limit format")
	}
	return ResourcePair{Request: strings.TrimSpace(parts[0]), Limit: strings.TrimSpace(parts[1])}, nil
}
