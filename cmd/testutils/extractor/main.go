package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

func main() {
	filePath := flag.String("p", "", "file path.")
	flag.Parse()

	bytes, err := os.ReadFile(*filePath)
	if err != nil {
		panic(fmt.Sprintf("Could not read file: %v with error %v", *filePath, err))
	}
	outerRegex := regexp.MustCompile(`Deployed AliceNetFactory at address: (.*),.*`)
	matchedOuter := outerRegex.FindAllSubmatch(bytes, -1)
	fmt.Printf("%s", matchedOuter[0][1])
}
