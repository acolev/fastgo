package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"fastgo/internal/config"
	"fastgo/internal/infra/database"
	"fastgo/internal/infra/database/seeds"
	"fastgo/internal/infra/database/seeds/dev"
	"fastgo/internal/infra/database/seeds/install"
	"fastgo/internal/shared/logger"
)

const (
	defaultSeed      = 42
	defaultSeedCount = 10
)

type options struct {
	command string
	target  string
	fresh   bool
	seed    uint64
	count   int
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		logger.Error("seeds failed", "error", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "list" {
		return
	}

	logger.Info("seeds completed")
}

func run(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}

	if opts.command == "list" {
		for _, name := range seedNames() {
			fmt.Println(name)
		}

		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger.Init(cfg.APP_NAME, cfg.APP_ENV)

	seeders, err := seedersFor(opts, cfg.APP_ENV)
	if err != nil {
		return err
	}

	if err := database.Init(cfg); err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("database shutdown failed", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return seeds.NewRunner(database.DB()).Run(ctx, seeders, opts.fresh)
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, errors.New("seed command is required: use list, run <name>, install, or dev")
	}

	opts := options{command: args[0]}
	flagArgs := args[1:]
	if opts.command == "run" {
		if len(flagArgs) == 0 {
			return options{}, errors.New("seed name is required: use run <name>")
		}

		opts.target = flagArgs[0]
		flagArgs = flagArgs[1:]
	}

	flags := flag.NewFlagSet("seed "+opts.command, flag.ContinueOnError)
	flags.BoolVar(&opts.fresh, "fresh", false, "reset group data before seeding")
	flags.Uint64Var(&opts.seed, "seed", defaultSeed, "faker seed for reproducible dev data")
	flags.IntVar(&opts.count, "count", defaultSeedCount, "number of generated dev records")

	if err := flags.Parse(flagArgs); err != nil {
		return options{}, err
	}

	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}

	return opts, nil
}

func seedersFor(opts options, appEnv string) ([]seeds.Seeder, error) {
	switch opts.command {
	case "install":
		if opts.fresh {
			return nil, errors.New("install seeds cannot be reset")
		}

		return install.Seeders(), nil
	case "dev":
		if !isDevelopment(appEnv) {
			return nil, errors.New("dev seeds can run only when APP_ENV=development")
		}

		return dev.Seeders(opts.seed, opts.count)
	case "run":
		return namedSeeders(opts.target, opts, appEnv)
	default:
		return nil, fmt.Errorf("unknown seed command %q: use list, run <name>, install, or dev", opts.command)
	}
}

func namedSeeders(name string, opts options, appEnv string) ([]seeds.Seeder, error) {
	switch name {
	case dev.NumbersName:
		if !isDevelopment(appEnv) {
			return nil, errors.New("dev seeds can run only when APP_ENV=development")
		}

		numberSeed, err := dev.NewNumbers(opts.seed, opts.count)
		if err != nil {
			return nil, err
		}

		return []seeds.Seeder{numberSeed}, nil
	default:
		return nil, fmt.Errorf("unknown seed %q: use seed list to see available seeds", name)
	}
}

func seedNames() []string {
	names := install.Names()
	return append(names, dev.Names()...)
}

func isDevelopment(appEnv string) bool {
	return strings.EqualFold(strings.TrimSpace(appEnv), "development")
}
