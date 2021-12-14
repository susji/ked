package util

func TruncateLine(rs []rune, width int, pad rune) []rune {
	if width <= 0 {
		return []rune("")
	}
	if len(rs) > width {
		i := len(rs) - width + 1
		rs = []rune(string(pad) + string(rs[i:]))
	}
	return rs
}
