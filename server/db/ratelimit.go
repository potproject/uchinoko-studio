package db

import (
	"context"
	"time"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

var timeNow = time.Now

func rateLimitFromRow(row sqlcgen.RateLimit) data.RateLimit {
	return data.RateLimit{
		Day: data.RateLimitType{
			LastUpdate: row.DayLastUpdate,
			Request:    row.DayRequest,
			Token:      row.DayToken,
		},
		Hour: data.RateLimitType{
			LastUpdate: row.HourLastUpdate,
			Request:    row.HourRequest,
			Token:      row.HourToken,
		},
		Minute: data.RateLimitType{
			LastUpdate: row.MinuteLastUpdate,
			Request:    row.MinuteRequest,
			Token:      row.MinuteToken,
		},
	}
}

func newRateLimitParams(id string, rateLimit data.RateLimit) sqlcgen.UpsertRateLimitParams {
	return sqlcgen.UpsertRateLimitParams{
		ID:               id,
		DayLastUpdate:    rateLimit.Day.LastUpdate,
		DayRequest:       rateLimit.Day.Request,
		DayToken:         rateLimit.Day.Token,
		HourLastUpdate:   rateLimit.Hour.LastUpdate,
		HourRequest:      rateLimit.Hour.Request,
		HourToken:        rateLimit.Hour.Token,
		MinuteLastUpdate: rateLimit.Minute.LastUpdate,
		MinuteRequest:    rateLimit.Minute.Request,
		MinuteToken:      rateLimit.Minute.Token,
	}
}

func rateLimitInit() data.RateLimit {
	now := timeNow()
	return data.RateLimit{
		Day: data.RateLimitType{
			LastUpdate: now.Format("20060102"),
			Request:    0,
			Token:      0,
		},
		Hour: data.RateLimitType{
			LastUpdate: now.Format("2006010215"),
			Request:    0,
			Token:      0,
		},
		Minute: data.RateLimitType{
			LastUpdate: now.Format("200601021504"),
			Request:    0,
			Token:      0,
		},
	}
}

func normalizeRateLimit(now time.Time, rateLimit data.RateLimit) data.RateLimit {
	dayPrefix := now.Format("20060102")
	hourPrefix := now.Format("2006010215")
	minutePrefix := now.Format("200601021504")

	if rateLimit.Day.LastUpdate != dayPrefix {
		rateLimit.Day.LastUpdate = dayPrefix
		rateLimit.Day.Request = 0
		rateLimit.Day.Token = 0
	}

	if rateLimit.Hour.LastUpdate != hourPrefix {
		rateLimit.Hour.LastUpdate = hourPrefix
		rateLimit.Hour.Request = 0
		rateLimit.Hour.Token = 0
	}

	if rateLimit.Minute.LastUpdate != minutePrefix {
		rateLimit.Minute.LastUpdate = minutePrefix
		rateLimit.Minute.Request = 0
		rateLimit.Minute.Token = 0
	}

	return rateLimit
}

type rateLimitQuerier interface {
	GetRateLimit(ctx context.Context, id string) (sqlcgen.RateLimit, error)
	UpsertRateLimit(ctx context.Context, arg sqlcgen.UpsertRateLimitParams) error
}

func getRateLimit(queryer rateLimitQuerier, id string) (data.RateLimit, error) {
	row, err := queryer.GetRateLimit(context.Background(), id)
	if isNotFound(err) {
		return rateLimitInit(), nil
	}
	if err != nil {
		return data.RateLimit{}, err
	}

	return rateLimitFromRow(row), nil
}

func putRateLimit(exec rateLimitQuerier, id string, config data.RateLimit) error {
	return exec.UpsertRateLimit(context.Background(), newRateLimitParams(id, config))
}

func PutRateLimitSnapshot(id string, config data.RateLimit) error {
	return putRateLimit(queries, id, config)
}

func AddRateLimit(id string, request int64, token int64) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQueries := queries.WithTx(tx)

	rateLimit, err := getRateLimit(txQueries, id)
	if err != nil {
		return err
	}

	rateLimit = normalizeRateLimit(timeNow(), rateLimit)
	rateLimit.Day.Request += request
	rateLimit.Hour.Request += request
	rateLimit.Minute.Request += request
	rateLimit.Day.Token += token
	rateLimit.Hour.Token += token
	rateLimit.Minute.Token += token

	if err := putRateLimit(txQueries, id, rateLimit); err != nil {
		return err
	}

	return tx.Commit()
}

func RateLimitIsAllowed(id string, limit data.CharacterConfigChatLimit) (bool, error) {
	rateLimit, err := getRateLimit(queries, id)
	if err != nil {
		return false, err
	}

	rateLimit = normalizeRateLimit(timeNow(), rateLimit)

	if rateLimit.Day.Request > limit.Day.Request && limit.Day.Request != 0 {
		return false, nil
	}
	if rateLimit.Hour.Request > limit.Hour.Request && limit.Hour.Request != 0 {
		return false, nil
	}
	if rateLimit.Minute.Request > limit.Minute.Request && limit.Minute.Request != 0 {
		return false, nil
	}
	if rateLimit.Day.Token > limit.Day.Token && limit.Day.Token != 0 {
		return false, nil
	}
	if rateLimit.Hour.Token > limit.Hour.Token && limit.Hour.Token != 0 {
		return false, nil
	}
	if rateLimit.Minute.Token > limit.Minute.Token && limit.Minute.Token != 0 {
		return false, nil
	}

	return true, nil
}
