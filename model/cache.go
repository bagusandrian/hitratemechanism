package model

import (
	"context"
	"time"
)

type (
	RequestCheck struct {
		Keys         []string
		ThresholdRPS int64
		TTLCache     time.Duration
	}
	Response struct {
		ResponseTime   string                   `json:"response_time"`
		DataKeys       map[string]DataTimeTrend `json:"data_time_trend"`
		Error          error                    `json:"error"`
		SuccessMessage string                   `json:"success_message"`
		Ctx            context.Context
	}
	DataTimeTrend struct {
		ReachThresholdRPS bool            `json:"reach_threshold_rps"`
		TimeTrend         map[int64]int64 `json:"time_trend"`
		EstimateRPS       int64           `json:"estimate_rps"`
		ThresholdRPS      int64           `json:"threshold_rps"`
		LimitTrend        int             `json:"limit_trend"`
	}
)
