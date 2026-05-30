package main

import (
	"strings"
	"testing"
)

func TestSeedersForRejectsDevOutsideDevelopment(t *testing.T) {
	_, err := seedersFor(options{command: "dev", seed: defaultSeed, count: defaultSeedCount}, "production")
	if err == nil || !strings.Contains(err.Error(), "only when APP_ENV=development") {
		t.Fatalf("expected development-only error, got %v", err)
	}
}

func TestSeedersForAllowsFreshInDevelopment(t *testing.T) {
	seeders, err := seedersFor(options{command: "dev", fresh: true, seed: defaultSeed, count: defaultSeedCount}, "development")
	if err != nil {
		t.Fatalf("seedersFor returned error: %v", err)
	}

	if len(seeders) != 1 {
		t.Fatalf("expected one dev seeder, got %d", len(seeders))
	}
}

func TestSeedersForRejectsInstallFresh(t *testing.T) {
	_, err := seedersFor(options{command: "install", fresh: true}, "development")
	if err == nil || !strings.Contains(err.Error(), "cannot be reset") {
		t.Fatalf("expected install reset error, got %v", err)
	}
}

func TestParseOptionsRunNamedSeed(t *testing.T) {
	opts, err := parseOptions([]string{"run", "dev.numbers", "--fresh", "--count", "10"})
	if err != nil {
		t.Fatalf("parseOptions returned error: %v", err)
	}

	if opts.command != "run" || opts.target != "dev.numbers" || !opts.fresh || opts.count != 10 {
		t.Fatalf("unexpected options: %+v", opts)
	}
}

func TestSeedNamesIncludesDevNumbers(t *testing.T) {
	names := seedNames()
	if len(names) != 1 || names[0] != "dev.numbers" {
		t.Fatalf("unexpected seed names: %v", names)
	}
}
