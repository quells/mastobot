package dbmigrations

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type logger struct{}

func (l logger) handle(e *zerolog.Event, v ...interface{}) {
	switch len(v) {
	case 0:
		e.Msg("no message")
	case 1:
		switch arg := v[0].(type) {
		case string:
			e.Msg(arg)
		case error:
			err := arg
			e.Err(err).Msg(err.Error())
		default:
			e.Msgf("%+v", arg)
		}
	default:
		e.Msgf("%+v", v)
	}
}

func (l logger) Fatal(v ...interface{}) {
	l.handle(log.Fatal(), v...)
}

func (l logger) Fatalf(format string, v ...interface{}) {
	log.Fatal().Msgf(format, v...)
}

func (l logger) Print(v ...interface{}) {
	l.handle(log.Debug(), v...)
}

func (l logger) Println(v ...interface{}) {
	l.Print(v...)
}

func (l logger) Printf(format string, v ...interface{}) {
	log.Debug().Msgf(format, v...)
}
