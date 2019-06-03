package gerbil

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func idxInSlice(a string, list []string) int {
	for idx, b := range list {
		if b == a {
			return idx
		}
	}
	return -1
}

func removeSliceStr(s []string, i int) []string {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}
