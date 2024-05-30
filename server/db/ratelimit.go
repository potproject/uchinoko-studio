package db

import (
	"encoding/json"
	"time"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/syndtr/goleveldb/leveldb"
)

func rateLimitInit() data.RateLimit {
	now := time.Now()
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

const rateLimitPrefix = "rate_limit_"

func getRateLimit(id string) (data.RateLimit, error) {
	key := []byte(rateLimitPrefix + id)
	value, err := get(key)
	if err == leveldb.ErrNotFound {
		return rateLimitInit(), nil
	} else if err != nil {
		return data.RateLimit{}, err
	}
	var config data.RateLimit
	err = json.Unmarshal(value, &config)
	if err != nil {
		return data.RateLimit{}, err
	}
	return config, nil
}

func putRateLimit(id string, config data.RateLimit) error {
	key := []byte(rateLimitPrefix + id)
	value, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return put(key, value)
}

func AddRateLimit(id string, request int64, token int64) error {
	now := time.Now()
	dayPrefix := now.Format("20060102")
	hourPrefix := now.Format("2006010215")
	minutePrefix := now.Format("200601021504")

	rateLimit, err := getRateLimit(id)
	if err != nil {
		return err
	}

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

	rateLimit.Day.Request += request
	rateLimit.Hour.Request += request
	rateLimit.Minute.Request += request

	rateLimit.Day.Token += token
	rateLimit.Hour.Token += token
	rateLimit.Minute.Token += token

	err = putRateLimit(id, rateLimit)
	if err != nil {
		return err
	}
	return nil
}

func RateLimitIsAllowed(id string, limit data.CharacterConfigChatLimit) (bool, error) {
	rateLimit, err := getRateLimit(id)
	if err != nil {
		return false, err
	}

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
