package database

import (
	"encoding/binary"
	"math"
	"time"
)

func getIndexFromLongitude(long float64) int64 {
	posIdx := int64(math.Trunc(long))
	return posIdx - posIdx%longPerTable
}

func truncateTime(timestamp int64) int64 {
	return time.Unix(timestamp, 0).Truncate(time.Hour * 24).Unix()
}

type NotFoundError struct {
	msg string
}

func (e NotFoundError) Error() string {
	return e.msg
}

func floatToBytes(num float64) [8]byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(num))
	return buf
}

func abs(num int64) int64 {
	if num < 0 {
		return -num
	}
	return num
}
