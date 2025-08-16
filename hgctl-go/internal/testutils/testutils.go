package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	L1Web3SignerUrl = "http://localhost:9100"
	L2Web3SignerUrl = "http://localhost:9101"
)

func GetProjectRootPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	startingPath := ""
	iterations := 0
	for {
		if iterations > 10 {
			panic("Could not find project root path")
		}
		iterations++
		p, err := filepath.Abs(fmt.Sprintf("%s/%s", wd, startingPath))
		if err != nil {
			panic(err)
		}

		match := regexp.MustCompile(`\/hourglass-monorepo(.+)?\/ponos$`)

		if match.MatchString(p) {
			return p
		}
		startingPath = startingPath + "/.."
	}
}
