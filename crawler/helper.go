package crawler

import (
	"hash/fnv"
)

func stringToUnsignedInteger32Hash(str string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(str))
	return h.Sum32()
}
