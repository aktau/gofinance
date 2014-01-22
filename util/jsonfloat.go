package util

import (
	"strconv"
)

/* use this type instead of plain float64 if you have a JSON stream which
* sometimes sends quoted floats and sometimes null's for the same field */
type NullFloat64 float64

func (n *NullFloat64) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		/* ignore and keep the default value */
		return nil
	}
	/* yes, really ugly, get rid of the quotes and convert */
	f, err := strconv.ParseFloat(string(b[1:len(b)-1]), 64)
	*n = NullFloat64(f)
	return err
}
