package www

import (
	"crypto/sha256"

	"github.com/RyuaNerin/ShareMe/share"
)

func genAES(pw string) (key []byte, iv []byte) {
	pwb := share.ToBytes(pw)

	hash := sha256.New()

	hash.Write(pwb)
	key = hash.Sum(nil)

	hash.Write(pwb)
	iv = hash.Sum(nil)

	return
}
