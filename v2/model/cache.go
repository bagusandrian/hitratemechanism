package model

type (
	RequestCheck struct {
		Key          string
		ThresholdRPS int64
	}
	Response struct {
		ResponseTime   string `json:"response_time"`
		DataTimeTrend  `json:"data_time_trend"`
		Error          error  `json:"error"`
		SuccessMessage string `json:"success_message"`
	}
	DataTimeTrend struct {
		HasCache     bool            `json:"has_cache"`
		TimeTrend    map[int64]int64 `json:"time_trend"`
		EstimateRPS  int64           `json:"estimate_rps"`
		ThresholdRPS int64           `json:"threshold_rps"`
		LimitTrend   int             `json:"limit_trend"`
	}
)
