package main

// Parser that spilts strings but keeps the seperators
type Parser struct{}

// Parse the string keeping the seperators
func (o *Parser) Parse(s string, sep rune) []string {
	list := make([]string, 0)

	lastIndex := 0
	for i, v := range s {
		if v == sep {
			if i == 0 {
				list = append(list, s[i:i+1])
			} else {
				list = append(list, s[lastIndex:i])
				list = append(list, s[i:i+1])
			}
			lastIndex = i + 1
		}
	}

	if lastIndex != len(s) {
		list = append(list, s[lastIndex:len(s)])
	}

	return list
}
