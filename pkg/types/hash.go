package types

import (
	"crypto/sha256"
	"encoding/hex"
)

// GetHashSuffix generates a short deterministic hash from a function name and its type arguments.
// Used to prevent name collisions in C when monomorphizing generic functions.
func GetHashSuffix(base string, typeArgs []NRType) string {
	h := sha256.New()
	h.Write([]byte(base))
	for _, ta := range typeArgs {
		h.Write([]byte(","))
		h.Write([]byte(ta.Name()))
	}
	return hex.EncodeToString(h.Sum(nil))[:8]
}

// GetParamSigHash generates a short deterministic hash from the names of a function's parameters.
// Used to disambiguate overloaded generic methods that share the same receiver type args.
// E.g. 'operator+(other: Box[T])' vs 'operator+(scalar: T)' when monomorphizing for Box[i32].
func GetParamSigHash(paramTypeNames []string) string {
	h := sha256.New()
	for _, name := range paramTypeNames {
		h.Write([]byte(","))
		h.Write([]byte(name))
	}
	res := hex.EncodeToString(h.Sum(nil))[:8]
	return res
}
