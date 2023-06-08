package sqlxx

import "unicode"

func LowerCase(s string) string {
	if s == "" {
		return s
	}
	rawStr := []rune(s)
	var newStr []rune
	preupper := true
	for idx, r := range rawStr {
		if unicode.IsUpper(r) {
			if idx > 0 && rawStr[idx-1] != '.' && !preupper {
				newStr = append(newStr, '_', unicode.ToLower(r))
				preupper = true
			} else {
				newStr = append(newStr, unicode.ToLower(r))
			}
		} else {
			preupper = false
			newStr = append(newStr, r)
		}
	}
	return string(newStr)
}
func BigCamelCase(s string) string {
	if s == "" {
		return s
	}
	rawStr := []rune(s)
	var newStr []rune
	toUpper := false
	for idx, r := range rawStr {
		if idx == 0 {
			newStr = append(newStr, unicode.ToUpper(r))
			continue
		}
		if r == '_' {
			toUpper = true
			continue
		}
		if r == '.' {
			newStr = append(newStr, r)
			toUpper = true
			continue
		}

		if toUpper {
			newStr = append(newStr, unicode.ToUpper(r))
			toUpper = false
		} else {
			newStr = append(newStr, r)
		}
	}
	return string(newStr)

}
func SmallCamelCase(s string) string {
	if s == "" {
		return s
	}
	rawStr := []rune(s)
	var newStr []rune
	toLower := false
	for idx, r := range rawStr {
		if idx == 0 {
			newStr = append(newStr, unicode.ToLower(r))
			continue
		}
		if r == '_' {
			toLower = true
			continue
		}
		if r == '.' {
			newStr = append(newStr, r)
			toLower = true
			continue
		}
		if toLower {
			newStr = append(newStr, unicode.ToLower(r))
			toLower = false
		} else {
			newStr = append(newStr, r)
		}
	}
	return string(newStr)
}
