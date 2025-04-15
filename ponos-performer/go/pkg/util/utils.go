package util

import "encoding/base64"

func DecodeBase64String(s string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func EncodeBase64String(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
