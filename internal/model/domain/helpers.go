package domain

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimeToProtobuf converts time.Time to protobuf timestamp
func TimeToProtobuf(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// TimePtr returns a pointer to the time value
func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
