package util

func Fmaybe(f *float64) float64 {
	if f != nil {
		return *f
	} else {
		return 0
	}
}
