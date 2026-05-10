package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mohamediag/infra-homelab-k8s-terraform-ansible-cilium/tools/generator/internal/emit"
	"github.com/mohamediag/infra-homelab-k8s-terraform-ansible-cilium/tools/generator/internal/schema"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("generator", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	servicePath := fs.String("service", "", "path to service.yaml")
	gitopsRepo := fs.String("gitops-repo", ".", "path to GitOps repository")
	imageTag := fs.String("image-tag", "", "image reference to write into generated App XRs")
	env := fs.String("env", "", "single environment to emit: dev, staging, or prod")
	allEnvs := fs.Bool("all-envs", false, "emit all environments defined in service.yaml")
	dryRun := fs.Bool("dry-run", false, "print generated files instead of writing them")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *servicePath == "" {
		return errors.New("--service is required")
	}
	svc, err := schema.Load(*servicePath)
	if err != nil {
		return err
	}
	if err := svc.Validate(*imageTag); err != nil {
		return err
	}

	envs, err := schema.SelectEnvs(svc.Environments, *env, *allEnvs)
	if err != nil {
		return err
	}

	repoRoot, err := filepath.Abs(*gitopsRepo)
	if err != nil {
		return fmt.Errorf("resolve --gitops-repo: %w", err)
	}

	files, err := emit.Files(svc, emit.Options{
		GitOpsRepo: repoRoot,
		ImageTag:   *imageTag,
		Envs:       envs,
	})
	if err != nil {
		return err
	}

	if *dryRun {
		for _, f := range files {
			fmt.Printf("--- %s\n%s", f.Path, f.Content)
		}
		return nil
	}

	if err := emit.Write(files); err != nil {
		return err
	}

	fmt.Printf("wrote %d files:\n", len(files))
	for _, f := range files {
		fmt.Printf("- %s\n", emit.RelativePath(repoRoot, f.Path))
	}

	return nil
}
