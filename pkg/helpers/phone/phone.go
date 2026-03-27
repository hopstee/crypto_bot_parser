package phone

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidPhone = errors.New("invalid russian phone number")
	digitsOnlyRe    = regexp.MustCompile(`\D`)
)

func NormalizeRUPhone(input string) (string, error) {
	digits := digitsOnlyRe.ReplaceAllString(input, "")

	if len(digits) == 10 {
		digits = "7" + digits
	}

	if len(digits) != 11 {
		return "", ErrInvalidPhone
	}

	if !strings.HasPrefix(digits, "7") && !strings.HasPrefix(digits, "8") {
		return "", ErrInvalidPhone
	}

	if digits[1] != '9' {
		return "", ErrInvalidPhone
	}

	digits = "7" + digits[1:]

	return digits, nil
}
