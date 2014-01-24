package util

func MapStr(mapping func(string) string, xs []string) []string {
	mxs := make([]string, 0, len(xs))
	for _, s := range xs {
		mxs = append(mxs, mapping(s))
	}
	return mxs
}
