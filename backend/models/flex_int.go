package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"strconv"
)

// FlexInt unmarshals JSON numbers (and numeric strings) into int by truncating toward zero.
// This avoids PostgreSQL errors when a fractional value (e.g. 9.9) is sent for an integer column.
type FlexInt int

func (f *FlexInt) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = 0
		return nil
	}
	var v float64
	if err := json.Unmarshal(data, &v); err == nil {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("invalid number for integer field")
		}
		*f = FlexInt(int(math.Trunc(v)))
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return errors.New("invalid number for integer field")
	}
	*f = FlexInt(int(math.Trunc(v)))
	return nil
}
