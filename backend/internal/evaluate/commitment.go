package evaluate

import (
	"encoding/hex"
)

func commitmentFromHash(hash string) string {
	if len(hash) >= 16 {
		return "0x" + hash[:16]
	}
	return "0x" + hex.EncodeToString([]byte(hash))
}
