package util

import (
	"encoding/json"
	"time"
)

var (
	FmtMonthDay      = "Jan _2"    /* yahoo finance often returns these types of dates */
	FmtDayTMonthYear = "_2-Jan-06" /* yahoo finance is quite inconsisten sometimes */
	FmtYearMonthDay  = "2006-01-02"
)

type JsonTime time.Time

/* shortime */
type MonthDay JsonTime

func (jt *MonthDay) UnmarshalJSON(data []byte) error {
	dt := (*JsonTime)(jt)
	return dt.JsonParse(data, FmtMonthDay, FmtDayTMonthYear)
}

func (jt *MonthDay) GetTime() time.Time {
	return (time.Time)(*jt)
}

type YearMonthDay JsonTime

func (jt *YearMonthDay) UnmarshalJSON(data []byte) error {
	dt := (*JsonTime)(jt)
	return dt.JsonParse(data, FmtYearMonthDay)
}

func (jt *YearMonthDay) GetTime() time.Time {
	return (time.Time)(*jt)
}

func (dt *JsonTime) JsonParse(data []byte, formats ...string) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	/* try all formats, from first to last */
	var err error
	for _, format := range formats {
		t, err := time.Parse(format, s)

		/* if one works, quit */
		if err == nil {
			*dt = (JsonTime)(t)
			return nil
		}
	}

	/* all formats failed, return error */
	return err
}
