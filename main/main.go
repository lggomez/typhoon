package main

import (
	"flag"
	"fmt"
	"github.com/lggomez/typhoon"
	"os"
)

func main() {
	pathArgPtr := flag.String("dir", "", "Path containing the source to analyze. If none, will os os.Getwd()")
	distanceArgPtr := flag.Int("dist", 2, "Levenshtein-Damerau distance threshold")
	flag.Parse()

	sourcePath := ""
	if *pathArgPtr == "" {
		sourcePath, _ = os.Getwd()
	} else {
		sourcePath = *pathArgPtr
	}

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		println("Invalid argument. A valid directory path is required: " + err.Error())
	}

	fmt.Printf("Starting query with Levenshtein-Damerau=%d\n", *distanceArgPtr)

	tree, queries := typhoon.IndexSourcesFromPath(&sourcePath)
	matches := typhoon.GetApproximateMatches(tree, queries, *distanceArgPtr)

	for k, v := range matches {
		fmt.Println("- query:" + k)
		for _, n := range v {
			pos := n.Position.String()
			fmt.Printf("\n\tapproximate match in: %s (value.ToLower:%s)", pos, n.Word)
		}
		fmt.Println("")
	}
}
