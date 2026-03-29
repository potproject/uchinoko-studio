package db

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/potproject/uchinoko-studio/data"
)

var timeNow = time.Now

type rateLimitRow struct {
	ID               string `db:"id"`
	DayLastUpdate    string `db:"day_last_update"`
	DayRequest       int64  `db:"day_request"`
	DayToken         int64  `db:"day_token"`
	HourLastUpdate   string `db:"hour_last_update"`
	HourRequest      int64  `db:"hour_request"`
	HourToken        int64  `db:"hour_token"`
	MinuteLastUpdate string `db:"minute_last_update"`
	MinuteRequest    int64  `db:"minute_request"`
	MinuteToken      int64  `db:"minute_token"`
}

func (r rateLimitRow) toRateLimit() data.RateLimit {
	return data.RateLimit{
		Day: data.RateLimitType{
			LastUpdate: r.DayLastUpdate,
			Request:    r.DayRequest,
			Token:      r.DayToken,
		},
		Hour: data.RateLimitType{
			LastUpdate: r.HourLastUpdate,
			Request:    r.HourRequest,
			Token:      r.HourToken,
		},
		Minute: data.RateLimitType{
			LastUpdate: r.MinuteLastUpdate,
			Request:    r.MinuteRequest,
			Token:      r.MinuteToken,
		},
	}
}

func newRateLimitRow(id string, rateLimit data.RateLimit) rateLimitRow {
	return rateLimitRow{
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

func getRateLimit(queryer sqlx.Queryer, id string) (data.RateLimit, error) {
	var row rateLimitRow
	err := sqlx.Get(queryer, &row, "SELECT * FROM rate_limits WHERE id = ?", id)
	if isNotFound(err) {
		return rateLimitInit(), nil
	}
	if err != nil {
		return data.RateLimit{}, err
	}

	return row.toRateLimit(), nil
}

func putRateLimit(exec sqlx.Ext, id string, config data.RateLimit) error {
	row := newRateLimitRow(id, config)
	_, err := sqlx.NamedExec(exec, `
		INSERT INTO rate_limits (
			id,
			day_last_update,
			day_request,
			day_token,
			hour_last_update,
			hour_request,
			hour_token,
			minute_last_update,
			minute_request,
			minute_token
		) VALUES (
			:id,
			:day_last_update,
			:day_request,
			:day_token,
			:hour_last_update,
			:hour_request,
			:hour_token,
			:minute_last_update,
			:minute_request,
			:minute_token
		)
		ON CONFLICT(id) DO UPDATE SET
			day_last_update = excluded.day_last_update,
			day_request = excluded.day_request,
			day_token = excluded.day_token,
			hour_last_update = excluded.hour_last_update,
			hour_request = excluded.hour_request,
			hour_token = excluded.hour_token,
			minute_last_update = excluded.minute_last_update,
			minute_request = excluded.minute_request,
			minute_token = excluded.minute_token
	`, row)
	return err
}

func PutRateLimitSnapshot(id string, config data.RateLimit) error {
	return putRateLimit(db, id, config)
}

func AddRateLimit(id string, request int64, token int64) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rateLimit, err := getRateLimit(tx, id)
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

	if err := putRateLimit(tx, id, rateLimit); err != nil {
		return err
	}

	return tx.Commit()
}

func RateLimitIsAllowed(id string, limit data.CharacterConfigChatLimit) (bool, error) {
	rateLimit, err := getRateLimit(db, id)
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
