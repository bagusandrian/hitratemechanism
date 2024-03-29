# Hitrate Mechanism Library with Client-Side Cache and Real-Time RPS Estimation

## Overview

Welcome to the Hitrate Mechanism Library, a powerful tool for optimizing Redis performance by combining client-side caching and real-time Requests Per Second (RPS) estimation. This library intelligently decides when to cache data locally, preventing memory bloat and improving overall system efficiency.

## Features

- **Real-Time RPS Estimation:** Dynamically calculate RPS to make informed caching decisions.
- **Configurable Parameters:** Easily adjust parameters such as time trend length, threshold RPS, and TTL for key memory checking.

## Getting Started

Follow these steps to integrate the library into your Redis-based application:

1. Write code on your golang code: 
	```golang
	import (
    "log"
    hrm "github.com/bagusandrian/hitratemechanism"
	hrmModel "github.com/bagusandrian/hitratemechanism/model"
	)

	func main() {
		uHrm := hrm.New(&hrmModel.HitRateMechanism{
			Config: hrmModel.Config{
				PrefixKey:         "hrm",
				DefaultExpiration: 60 * time.Second,
				LimitTrend:        10,
			},
		})
		resp := r.uHrm.HandlerCache.CacheValidateTrend(ctx, hrmModel.RequestCheck{
			Keys:         keys,
			ThresholdRPS: 2,
			TTLCache:     60 * time.Second,
		})
		log.Printf("response of HRM: %+v\n", resp)
	//	{
	//		"response_time": "20ms",
	//		"data_time_trend": {
	//			"reach_threshold_rps": true,
	//			"time_trend": {
	//				"1": {
	//					0: 1234567,
	//					1: 1234580,
	//					2: 1234600,
	//					3: 1234620,
	//					4: 1234640
	//				},
	//				"2": {
	//					0: 1234567,
	//					1: 1234580,
	//					2: 1234600,
	//					3: 1234620,
	//					4: 1234640
	//				},
	//		},
	//			"estimate_rps": 100,
	//			"threshold_rps": 2,
	//			"limit_trend": 5
	//		}
	//	}
	}
	```
