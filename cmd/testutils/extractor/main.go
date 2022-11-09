package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	network := flag.String("n", "", "network name.")
	filePath := flag.String("p", "", "file path.")
	flag.Parse()

	bytes, err := os.ReadFile(*filePath)
	if err != nil {
		panic("Could nor read file")
	}
	outerRegex := regexp.MustCompile(fmt.Sprintf(`\[%s\]\ndefaultFactoryAddress = \".*\"\n`, *network))
	innerRegex := regexp.MustCompile(`defaultFactoryAddress = .*`)
	matchedOuter := outerRegex.FindAllSubmatch(bytes, -1)
	for i := 0; i < len(matchedOuter); i++ {
		var innerSentence string
		for j := 0; j < len(matchedOuter[i]); j++ {
			innerSentence += string(matchedOuter[i][j])
		}
		matchedInner := innerRegex.FindAllSubmatch([]byte(innerSentence), -1)
		tempResult := matchedInner[0]
		var sentence string
		for j := 0; j < len(tempResult); j++ {
			sentence += string(tempResult[j])
		}
		splitBySpace := strings.Split(sentence, " ")
		fmt.Printf("%s", splitBySpace[len(splitBySpace)-1])
	}
}
