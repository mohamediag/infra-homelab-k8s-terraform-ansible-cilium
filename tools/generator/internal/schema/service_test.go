package schema

import "testing"

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
