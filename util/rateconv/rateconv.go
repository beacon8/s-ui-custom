// Package rateconv converts the client-facing Mbps limit to bytes per second.
//
// The conversion is deliberately done in int64. A plain int multiplication
// overflows on 32-bit targets for values above 17179 Mbps, and the resulting
// negative rate would make a limiter behave unpredictably.
package rateconv

import (
	"errors"
	"fmt"
)

const (
	// BytesPerMbps is 1,000,000 bits / 8.
	BytesPerMbps int64 = 125_000

	// MaxMbps is the largest integer Mbps value whose byte rate still fits in
	// a signed 32-bit integer. This makes the limit portable to 386/arm/wasm32:
	// 17179 * 125000 = 2,147,375,000; 17180 * 125000 overflows int32.
	MaxMbps int64 = (int64(1)<<31 - 1) / BytesPerMbps
)

var ErrTooLarge = fmt.Errorf("rate exceeds maximum of %d Mbps", MaxMbps)

// BytesPerSecond converts Mbps to bytes per second. It returns ErrTooLarge
// instead of wrapping when the value does not fit in 32 bits.
func BytesPerSecond(mbps int) (int64, error) {
	if mbps < 0 {
		return 0, errors.New("rate must not be negative")
	}
	if int64(mbps) > MaxMbps {
		return 0, ErrTooLarge
	}
	return int64(mbps) * BytesPerMbps, nil
}
