package cmd

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"sync"
	"time"
)

var (
	shutdownFuncs []func()
	shutdownOnce  = new(sync.Once)
)

func shutdown() {
	shutdownOnce.Do(func() {
		log.Debug().Msg("running shutdown funcs")
		for _, f := range shutdownFuncs {
			f()
		}
		log.Debug().Msg("finished shutdown funcs")
	})
}

func registerShutdown(f func()) {
	shutdownFuncs = append(shutdownFuncs, f)
}

func must(err error) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		shutdown()
		os.Exit(1)
	}
}

var (
	connStr  string
	instance string
	timeout  time.Duration

	v  bool
	vv bool
)

var rootCmd = &cobra.Command{
	Use:   "mastobot",
	Short: "Mastodon Bots",
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		registerShutdown(cancel)
		cmd.SetContext(ctx)

		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		if v {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
		if vv {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&connStr, "db", "mastobot.db", "sqlite database connection string")
	rootCmd.PersistentFlags().StringVar(&instance, "instance", "", "Mastodon (or compatible) instance to interact with")
	must(rootCmd.MarkPersistentFlagRequired("instance"))
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "Request timeout")

	rootCmd.PersistentFlags().BoolVar(&v, "v", false, "Log info level")
	rootCmd.PersistentFlags().BoolVar(&vv, "vv", false, "Log debug level")
}

func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	must(rootCmd.Execute())
	shutdown()
}
