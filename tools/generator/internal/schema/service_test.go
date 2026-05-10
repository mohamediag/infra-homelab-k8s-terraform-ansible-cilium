package schema

import (
	"path/filepath"
	"testing"
)

func TestSelectEnvs(t *testing.T) {
	envs := map[string]Environment{
		"staging": {},
		"dev":     {},
	}

	got, err := SelectEnvs(envs, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "dev" || got[1] != "staging" {
		t.Fatalf("unexpected env order: %#v", got)
	}

	got, err = SelectEnvs(envs, "dev", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "dev" {
		t.Fatalf("unexpected single env: %#v", got)
	}
}

func TestSelectEnvsRejectsUndefinedEnv(t *testing.T) {
	_, err := SelectEnvs(map[string]Environment{"dev": {}}, "prod", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseResourcePair(t *testing.T) {
	pair, err := ParseResourcePair("25m/100m")
	if err != nil {
		t.Fatal(err)
	}
	if pair.Request != "25m" || pair.Limit != "100m" {
		t.Fatalf("unexpected pair: %#v", pair)
	}

	if _, err := ParseResourcePair("25m"); err == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestAutoCommitDevDefault(t *testing.T) {
	svc := Service{}
	if !svc.AutoCommitDev() {
		t.Fatal("expected autoCommitDev to default to true")
	}
}

func TestAutoCommitDevExplicitFalse(t *testing.T) {
	autoCommitDev := false
	svc := Service{CI: &CI{AutoCommitDev: &autoCommitDev}}
	if svc.AutoCommitDev() {
		t.Fatal("expected explicit false to be honored")
	}
}

func TestAutoCommitDevDefaultsWithCIBlock(t *testing.T) {
	svc := Service{CI: &CI{Promotion: &Promotion{Staging: "commit", Prod: "pr"}}}
	if !svc.AutoCommitDev() {
		t.Fatal("expected omitted ci.autoCommitDev to default to true")
	}
}

func TestLoadAcceptsCI(t *testing.T) {
	svc, err := Load(filepath.Join("..", "..", "testdata", "basic", "service.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !svc.AutoCommitDev() {
		t.Fatal("expected ci.autoCommitDev=true from fixture")
	}
}

func TestPromotionModeDefaults(t *testing.T) {
	svc := Service{}

	mode, err := svc.PromotionMode("staging")
	if err != nil {
		t.Fatal(err)
	}
	if mode != "commit" {
		t.Fatalf("expected staging default commit, got %q", mode)
	}

	mode, err = svc.PromotionMode("prod")
	if err != nil {
		t.Fatal(err)
	}
	if mode != "pr" {
		t.Fatalf("expected prod default pr, got %q", mode)
	}
}

func TestPromotionModeOverrides(t *testing.T) {
	svc := Service{CI: &CI{Promotion: &Promotion{Staging: "pr", Prod: "commit"}}}

	mode, err := svc.PromotionMode("staging")
	if err != nil {
		t.Fatal(err)
	}
	if mode != "pr" {
		t.Fatalf("expected staging override pr, got %q", mode)
	}

	mode, err = svc.PromotionMode("prod")
	if err != nil {
		t.Fatal(err)
	}
	if mode != "commit" {
		t.Fatalf("expected prod override commit, got %q", mode)
	}
}

func TestPromotionModeRejectsDev(t *testing.T) {
	svc := Service{}
	if _, err := svc.PromotionMode("dev"); err == nil {
		t.Fatal("expected error")
	}
}
