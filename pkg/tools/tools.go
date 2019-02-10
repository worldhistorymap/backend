package tools

import "os"

var JwtSecretKey = []byte(GetEnv("jwt_secret_key", "aVerySecretKey"))
const (
	ScryptN = 32768
	ScryptR = 8
	ScryptP = 1
	ScryptKeyLen = 32
)

func GetEnv(key, defaultVal string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultVal
	}
	return key
}