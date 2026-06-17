package types

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetHashSuffix(baseName string, argTypes []NRType) string {
	h := sha256.New()
	h.Write([]byte(baseName))
	for _, arg := range argTypes {
		h.Write([]byte(","))
		if arg != nil {
			h.Write([]byte(arg.Name()))
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:8]
}
