package qr

import (
	"github.com/skip2/go-qrcode"
)

func GenerateQRPNG(data string, size int) ([]byte, error) {
	return qrcode.Encode(data, qrcode.Medium, size)
}
