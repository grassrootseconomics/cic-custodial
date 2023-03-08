package sub

import (
	"encoding/hex"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

// TODO: This should probably be used project wide
func checksum(address string) string {
	address = strings.ToLower(address)
	address = strings.Replace(address, "0x", "", 1)

	sha := sha3.NewLegacyKeccak256()
	sha.Write([]byte(address))
	hash := sha.Sum(nil)
	hashstr := hex.EncodeToString(hash)
	result := []string{"0x"}
	for i, v := range address {
		res, _ := strconv.ParseInt(string(hashstr[i]), 16, 64)
		if res > 7 {
			result = append(result, strings.ToUpper(string(v)))
			continue
		}
		result = append(result, string(v))
	}

	return strings.Join(result, "")
}
