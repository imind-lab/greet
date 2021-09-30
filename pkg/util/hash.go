/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package util

import "math"

var HashBase = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

func Base62encode(id int32) string {
	baseStr := ""
	for {
		if id <= 0 {
			break
		}
		i := id % 62
		baseStr += HashBase[i]
		id = (id - i) / 62
	}
	return baseStr
}
func Base62decode(base62 string) int {
	rs := 0
	length := len(base62)
	f := flip(HashBase)
	for i := 0; i < length; i++ {
		rs += f[string(base62[i])] * int(math.Pow(62, float64(i)))
	}
	return rs
}
func flip(s []string) map[string]int {
	f := make(map[string]int)
	for index, value := range s {
		f[value] = index
	}
	return f
}
