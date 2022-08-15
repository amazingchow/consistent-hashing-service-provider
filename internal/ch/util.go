package consistenthashing

import (
	"fmt"

	"github.com/amazingchow/consistent-hashing-service-provider/internal/hashlib"
)

func genLoc(uuid string, idx int) uint32 {
	return hashlib.FNV1av32(fmt.Sprintf("%s#%d", uuid, idx))
}

func prettyPrint(uuids []string) { // nolint
	maxLen := 0
	for _, uuid := range uuids {
		if len(uuid) > maxLen {
			maxLen = len(uuid)
		}
	}
	header := "All Physical Nodes"
	if len(header) > maxLen {
		maxLen = len(header)
	}
	dashes := make([]byte, maxLen+10)
	for i := 0; i < maxLen+10; i++ {
		dashes[i] = '-'
	}
	blanks := make([]byte, maxLen+10-len(header)-1)
	for i := 0; i < maxLen+10-len(header)-1; i++ {
		blanks[i] = ' '
	}
	fmt.Printf("+%s+\n", string(dashes))
	fmt.Printf("| %s%s|\n", header, blanks)
	fmt.Printf("+%s+\n", string(dashes))
	fills := make([][]byte, len(uuids))
	for i := 0; i < len(uuids); i++ {
		x := make([]byte, maxLen+10-len(uuids[i])-1)
		for j := 0; j < maxLen+10-len(uuids[i])-1; j++ {
			x[j] = ' '
		}
		fills[i] = x
	}
	for i := 0; i < len(uuids); i++ {
		fmt.Printf("| %s%s|\n", uuids[i], fills[i])
	}
	fmt.Printf("+%s+\n", string(dashes))
}
