package util

import (
	"encoding/json"
	"time"
)

var (
	FmtMonthDay     = "Jan 02"
	FmtYearMonthDay = "2006-01-02"
)

type JsonTime time.Time

/* shortime */
type MonthDay JsonTime

func (jt *MonthDay) UnmarshalJSON(data []byte) error {
	dt := (*JsonTime)(jt)
	return dt.JsonParse(data, FmtMonthDay)
}

type YearMonthDay JsonTime

func (jt *YearMonthDay) UnmarshalJSON(data []byte) error {
	dt := (*JsonTime)(jt)
	return dt.JsonParse(data, FmtYearMonthDay)
}

func (jt *MonthDay) GetTime() time.Time {
	return (time.Time)(*jt)
}

func (dt *JsonTime) JsonParse(data []byte, format string) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(format, s)
	if err != nil {
		return err
	}

	*dt = (JsonTime)(t)

	return nil
}
