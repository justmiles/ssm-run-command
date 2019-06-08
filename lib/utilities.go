package command

import (
	"fmt"
	"math/rand"
	"time"
)

func toString(p map[string][]*string) (s string) {
	for _, v := range p {
		for _, vv := range v {
			s = s + fmt.Sprintf("  %s\n", *vv)
		}
	}
	return s
}

func randomSeconds(i int) time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(rand.Intn(i)) * time.Second
}
