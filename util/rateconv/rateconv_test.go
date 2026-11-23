package rateconv

import (
	"errors"
	"strconv"
	"testing"
)

func TestBytesPerSecondBoundary(t *testing.T) {
	got, err := BytesPerSecond(17179)
	if err != nil {
		t.Fatal(err)
	}
	if got != 2_147_375_000 {
		t.Fatalf("17179 Mbps = %d, want 2147375000", got)
	}

	if _, err = BytesPerSecond(17180); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("17180 Mbps error = %v, want ErrTooLarge", err)
	}
	if _, err = BytesPerSecond(-1); err == nil {
		t.Fatal("negative rate must be rejected")
	}
}

func Test32BitTarget(t *testing.T) {
	if strconv.IntSize != 32 {
		t.Skipf("not a 32-bit target (int is %d bits)", strconv.IntSize)
	}
	got, err := BytesPerSecond(17179)
	if err != nil || got != 2_147_375_000 {
		t.Fatalf("32-bit 17179: got %d, %v", got, err)
	}
	if _, err = BytesPerSecond(17180); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("32-bit 17180 did not return ErrTooLarge: %v", err)
	}
}
