package tgauth

import (
	"encoding/json"
	"strings"
)

func ParseAuthResult(body string) (*TgAuthResult, bool, error) {
	start := strings.Index(body, "result: {")
	if start == -1 {
		return nil, false, nil
	}

	start += len("result: ")
	end := strings.Index(body[start:], "}, origin")
	if end == -1 {
		return nil, false, nil
	}

	raw := body[start : start+end+1]

	var r TgAuthResult
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, false, err
	}

	return &r, true, nil
}

func ObtainAccessHash(body string) string {
	if !strings.Contains(body, "confirm=1&hash=") {
		return ""
	}

	parts := strings.Split(body, "confirm=1&hash=")
	if len(parts) <= 1 {
		return ""
	}

	return strings.Split(parts[1], "'")[0]
}
