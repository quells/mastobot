package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	shutdownFuncs []func()
	shutdownOnce  = new(sync.Once)
)

func shutdown() {
	shutdownOnce.Do(func() {
		log.Debug().Msg("starting shutdown sequence")
		for _, f := range shutdownFuncs {
			f()
		}
		log.Debug().Msg("finished shutdown sequence")
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
	connStr string
	db      *sql.DB

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

		var err error
		db, err = sql.Open("sqlite3", connStr)
		must(err)
		registerShutdown(func() {
			log.Debug().Msg("closing database connection")
			if cErr := db.Close(); cErr != nil {
				log.Error().Err(err).Msg("failed to close database connection")
			}
		})
		must(db.Ping())
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&connStr, "db", "file:mastobot.db?_busy_timeout=5000&_journal_mode=WAL", "sqlite database connection string")
	rootCmd.PersistentFlags().StringVar(&instance, "instance", "", "Mastodon (or compatible) instance to interact with")
	must(rootCmd.MarkPersistentFlagRequired("instance"))
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "Request timeout")

	rootCmd.PersistentFlags().BoolVarP(&v, "log_info", "v", false, "Log info level")
	rootCmd.PersistentFlags().BoolVarP(&vv, "log_debug", "V", false, "Log debug level")
}

func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	must(rootCmd.Execute())
	shutdown()
}
