package main

import (
	"fmt"
	"github.com/lggomez/typhoon"
)

func main() {
	distanceArgPtr, sourcePath := typhoon.ParseArgs()
	fmt.Printf("Starting query with Levenshtein-Damerau=%d\n", *distanceArgPtr)
	tree, queries := typhoon.IndexSourcesFromPath(&sourcePath)
	matches := typhoon.GetApproximateMatches(tree, queries, *distanceArgPtr)

	for query, match := range *matches {
		fmt.Printf("\n- Selected string literal: %s", query)
		for _, resultInfo := range match {
			pos := resultInfo.Node.Position.String()
			fmt.Printf("\n\tapproximate match in: %s (value.ToLower:%s)", pos, resultInfo.AssociatedQueries)
		}
		fmt.Println("")
	}
}
