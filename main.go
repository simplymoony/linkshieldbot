package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/simplymoony/linkshieldbot/internal/tg"
	"github.com/simplymoony/linkshieldbot/internal/tg/poller"
)

const (
	progName = "linkshieldbot"
	usage    = `%s %s
	
Usage:
  %s [options]

Environment variables:
  BOT_TOKEN	your bot's API token (required)

Options:
  -config <path>
      path to config file
  -verbose
      emit verbose logs (has priority over config)
`
)

var version string // Initialized in init

type env struct {
	config
	*log.Logger
}

func (e *env) Verbosef(format string, v ...any) {
	if e.Verbose {
		e.Printf(format, v...)
	}
}

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		version = buildInfo.Main.Version
	} else {
		version = "(unknown)"
	}
}

func main() {
	var (
		cfgPathFlag string
		verboseFlag bool
	)

	fs := flag.NewFlagSet(progName, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Prevents flag package from printing things
	fs.StringVar(&cfgPathFlag, "config", "", "path to config file")
	fs.BoolVar(&verboseFlag, "verbose", false, "emit verbose logs (has priority over config)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(os.Stderr, usage, progName, version, os.Args[0])
			return
		}
		exitErr("Invalid options\nUse option -help for help")
	}

	var cfgPath string
	if cfgPathFlag != "" {
		cfgPath = cfgPathFlag
	} else {
		var err error
		if cfgPath, err = configPath(); err != nil {
			exitErr("Failed to determine or set-up your system's config directory: %v\n"+
				"Try providing a custom path using the -config flag",
				err,
			)
		}
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		exitErr("Missing BOT_TOKEN environment variable")
	}

	e := &env{
		config{
			PollerTimeout:  10,
			HandlerTimeout: 20,
		},
		log.New(os.Stderr, "", log.Ldate|log.Ltime),
	}

	nf, err := loadConfig(cfgPath, &e.config)
	if err != nil {
		if nf {
			exitErr("Failed to generate config file: %v", err)
		}
		exitErr("Failed to load config file: %v", err)
	}
	if nf {
		exitErr("Config file not found, but it was generated (path = %s)\n"+
			"Set it up to your needs and run me again once you're done",
			cfgPath,
		)
	}

	fs.Visit(func(f *flag.Flag) {
		if f.Name == "verbose" {
			e.Verbose = verboseFlag
		}
	})

	if len(e.Directives) == 0 {
		exitErr("No directives were found in config, a minimum of one directive is required")
	}

	e.Printf("Starting up..")
	e.Verbosef("Verbose logging is enabled!")
	e.Verbosef("Environment:\n"+
		"OS: %s\n"+
		"Version: %s\n"+
		"Config path: %s\n"+
		"PollerTimeout: %d\n"+
		"HandlerTimeout: %d\n"+
		"Loaded directives: %d",
		runtime.GOOS, version, cfgPath, e.PollerTimeout, e.HandlerTimeout, len(e.Directives),
	)

	ctx, stop := signalContext(e.Logger)
	defer stop()

	bot := tg.NewBot(botToken)

	me, err := bot.GetMe(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		e.Fatalf("**FATAL** Healthcheck request failed: %v\n"+
			"Please check your connectivity or that the entered Bot API token is correct",
			err,
		)
	}

	e.Printf("Running as %s (@%s)", me.FirstName, me.Username)

	err = poller.Poll(
		ctx, bot,
		time.Duration(e.PollerTimeout)*time.Second,
		time.Duration(e.HandlerTimeout)*time.Second,
		routeUpdate(e), handleError(e),
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		e.Fatalf("**FATAL** Error during polling: %v", err)
	}
}

func exitErr(format string, a ...any) {
	fmt.Printf("error: "+format+"\n", a...)
	os.Exit(1)
}

// Attempts to retrieve the system's standard config directory, creating a subfolder named after the
// application if not already existing.
func configPath() (string, error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	path = filepath.Join(path, progName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(path, 0755); err != nil {
			return "", err
		}
	}
	return filepath.Join(path, "config.toml"), nil
}

// Creates a context that is cancelled when an interrupt signal is received and logs a message with
// the given logger
func signalContext(logger *log.Logger) (ctx context.Context, stop context.CancelFunc) {
	ctx, stop = context.WithCancel(context.Background())

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		<-sigch
		logger.Printf("Interrupt signal received! Shutting down..")
		stop()
	}()

	return
}
