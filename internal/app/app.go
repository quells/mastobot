package app

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/quells/mastobot/internal/dbcontext"
	"github.com/rs/zerolog/log"
)

func Exists(ctx context.Context, instance, appName string) (exists bool, err error) {
	var query string
	var params []any
	query, params, err = goqu.
		Select("instance").
		From("apps").
		Where(goqu.Ex{
			"instance": instance,
			"app_name": appName,
		}).
		ToSQL()
	if err != nil {
		return
	}
	log.Debug().Msg(query)

	var db *sql.DB
	db, err = dbcontext.From(ctx)
	if err != nil {
		return
	}

	err = db.QueryRow(query, params...).Scan(&instance)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}

	exists = true
	return
}

func Register(ctx context.Context, instance, appName, appID, clientID, clientSecret string) (err error) {
	var stmt string
	var params []any
	stmt, params, err = goqu.
		Insert("apps").
		Cols("instance", "app_name", "app_id", "client_id", "client_secret").
		Vals(goqu.Vals{instance, appName, appID, clientID, clientSecret}).
		ToSQL()
	if err != nil {
		return
	}
	log.Debug().Msg(stmt)

	var db *sql.DB
	db, err = dbcontext.From(ctx)
	if err != nil {
		return
	}

	_, err = db.ExecContext(ctx, stmt, params...)
	if err != nil {
		return
	}

	return nil
}

func GetClientSecrets(ctx context.Context, instance, appName string) (clientID, clientSecret string, err error) {
	var query string
	var params []any
	query, params, err = goqu.
		Select("client_id", "client_secret").
		From("apps").
		Where(goqu.Ex{
			"instance": instance,
			"app_name": appName,
		}).
		ToSQL()
	if err != nil {
		return
	}
	log.Debug().Msg(query)

	var db *sql.DB
	db, err = dbcontext.From(ctx)
	if err != nil {
		return
	}

	err = db.QueryRow(query, params...).Scan(&clientID, &clientSecret)
	if err != nil {
		return
	}

	return
}

func UpdateAccessToken(ctx context.Context, instance, appName, token string) (err error) {
	var stmt string
	var params []any
	stmt, params, err = goqu.
		Update("apps").
		Set(goqu.Record{
			"access_token": token,
		}).
		Where(goqu.Ex{
			"instance": instance,
			"app_name": appName,
		}).
		ToSQL()
	if err != nil {
		return
	}
	log.Debug().Msg(stmt)

	var db *sql.DB
	db, err = dbcontext.From(ctx)
	if err != nil {
		return
	}

	_, err = db.ExecContext(ctx, stmt, params...)
	if err != nil {
		return
	}

	return
}

func GetAccessToken(ctx context.Context, instance, appName string) (token string, err error) {
	var query string
	var params []any
	query, params, err = goqu.
		Select("access_token").
		From("apps").
		Where(goqu.Ex{
			"instance": instance,
			"app_name": appName,
		}).
		ToSQL()
	if err != nil {
		return
	}
	log.Debug().Msg(query)

	var db *sql.DB
	db, err = dbcontext.From(ctx)
	if err != nil {
		return
	}

	err = db.QueryRow(query, params...).Scan(&token)
	if err != nil {
		return
	}

	return
}
