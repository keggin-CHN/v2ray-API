package xrayruntime

import "encoding/base64"

func decodeBase64Any(s string) string {
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		b, err := enc.DecodeString(s)
		if err == nil {
			return string(b)
		}
	}
	return ""
}
