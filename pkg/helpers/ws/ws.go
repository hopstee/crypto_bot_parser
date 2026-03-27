package ws

import (
	"encoding/base64"
	"errors"

	"github.com/gorilla/websocket"
)

func IsNormalWSClose(err error) bool {
	var ce *websocket.CloseError
	if errors.As(err, &ce) {
		switch ce.Code {
		case websocket.CloseNormalClosure,
			websocket.CloseGoingAway,
			websocket.CloseNoStatusReceived: // 1005
			return true
		}
	}
	return false
}

func DecodeIfBase64(msg []byte) ([]byte, error) {
	if len(msg) > 0 && msg[0] == 'M' {
		decoded, err := base64.StdEncoding.DecodeString(string(msg))
		if err == nil {
			return decoded, nil
		}
	}
	return msg, nil
}
