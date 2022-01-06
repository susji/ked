package util

import "strings"

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

func SplitRunesOnWidth(rs []rune, width int) [][]rune {
	if width <= 0 {
		return [][]rune{}
	}
	if len(rs) <= width {
		return [][]rune{rs}
	}

	ret := [][]rune{}
	for i := 0; i < len(rs); i += width {
		end := i + width
		if end < len(rs) {
			ret = append(ret, rs[i:end])
		} else {
			ret = append(ret, rs[i:])
		}
	}
	return ret
}

func Unescape(raw string) string {
	ret := strings.ReplaceAll(raw, `\a`, "\a")
	ret = strings.ReplaceAll(ret, `\b`, "\b")
	ret = strings.ReplaceAll(ret, `\t`, "\t")
	ret = strings.ReplaceAll(ret, `\n`, "\n")
	ret = strings.ReplaceAll(ret, `\v`, "\v")
	ret = strings.ReplaceAll(ret, `\r`, "\r")
	ret = strings.ReplaceAll(ret, `\f`, "\f")
	ret = strings.ReplaceAll(ret, `\r`, "\r")
	return ret
}
