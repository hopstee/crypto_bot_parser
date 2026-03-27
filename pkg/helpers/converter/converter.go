package converter

func FastAmount(s string) int64 {
	n := int64(0)
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			break
		}
		n = n*10 + int64(s[i]-'0')
	}
	return n
}
