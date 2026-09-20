package datetime

import (
	"errors"
	"strings"
)

var errInvalidFormat = errors.New(ErrInvalidFormat)

const (
	Flag             = "datetime"
	ErrInvalidFormat = "invalid format"

	formatYYMMDDhhmmss = "YYMMDDhhmmss"

	fYY   = "06"
	fYYY  = "006"
	fYYYY = "2006"
	fMM   = "01"
	fDD   = "02"
	fhh   = "15"
	fmm   = "04"
	fss   = "05"
)

func Parse(raw string) (string, error) {
	rawB := []byte(raw)
	if len(rawB) < len(formatYYMMDDhhmmss) {
		return "", errInvalidFormat
	}
	var b strings.Builder
	var year, month, day, hour, minute, second bool
	for i := 0; i < len(rawB); i++ {
		switch rawB[i] {
		case 'Y':
			if year {
				return "", errInvalidFormat
			}
			switch index(rawB[i:], 'Y') {
			case 1:
				b.WriteString(fYY)
				i++
			case 2:
				b.WriteString(fYYY)
				i += 2
			case 3:
				b.WriteString(fYYYY)
				i += 3
			default:
				return "", errInvalidFormat
			}
			year = true
		case 'M':
			if month || !nextToken(rawB[i:], 'M') {
				return "", errInvalidFormat
			}
			i++
			b.WriteString(fMM)
			month = true
		case 'D':
			if day || !nextToken(rawB[i:], 'D') {
				return "", errInvalidFormat
			}
			i++
			b.WriteString(fDD)
			day = true
		case 'h':
			if hour || !nextToken(rawB[i:], 'h') {
				return "", errInvalidFormat
			}
			i++
			b.WriteString(fhh)
			hour = true
		case 'm':
			if minute || !nextToken(rawB[i:], 'm') {
				return "", errInvalidFormat
			}
			i++
			b.WriteString(fmm)
			minute = true
		case 's':
			if second || !nextToken(rawB[i:], 's') {
				return "", errInvalidFormat
			}
			i++
			b.WriteString(fss)
			second = true
		case '-':
			b.WriteByte('-')
		case '_':
			b.WriteByte('_')
		default:
			return "", errInvalidFormat
		}
	}
	if !year || !month || !day || !hour || !minute || !second {
		return "", errInvalidFormat
	}
	return b.String(), nil
}

func index(s []byte, c byte) int {
	last := -1
	for i := range s {
		if s[i] == c {
			last++
			continue
		}
		return last
	}
	return last
}

func nextToken(s []byte, c byte) bool { return len(s) > 1 && s[1] == c }
