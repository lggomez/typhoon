package typhoon

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// String format placeholder regex
// Adapted from https://stackoverflow.com/a/29403060
var formatPlaceholderRegex, _ = regexp.Compile("%(?:\\x25\\x25)|(\\x25(?:(?:[1-9]\\d*)\\$|\\((?:[^\\)]+)\\))?(?:\\+)?(?: )?(?:\\#)?(?:0|'[^$])?(?:-)?(?:\\d+)?(?:\\.(?:\\d+))?(?:[vT%bcdoqxXUeEfFgGsqp]))")

func inspect(fileName string, tree *Tree, queries *[]string) {
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, fileName, nil, 0)
	if err != nil {
		panic(err)
	}

	// Import token tracking variables
	importStartPos := token.NoPos
	importEndPos := token.NoPos

	// Inspect the AST identify all string literals.
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			// Ignore import tokens
			if x.Tok == token.IMPORT {
				// Capture import token positions to filter literals later on
				importStartPos = x.Pos()
				importEndPos = x.End()
			}
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				// Prevent duplicates
				if !inSlice(x.Value, *queries) {
					// Ignore empty strings and string format placeholder literals
					if x.Value == "" || (formatPlaceholderRegex.FindString(x.Value) != "" && len(x.Value) <= 5) {
						return true
					}

					// Some import paths may reach this part. We filter them by position
					if importStartPos != token.NoPos && importEndPos != token.NoPos {
						if importStartPos <= x.Pos() && importEndPos >= x.End() {
							return true
						}
					}

					*queries = append(*queries, x.Value)
					position := fset.Position(n.Pos())
					tree.Add(x.Value, &position)
					//fmt.Printf("Adding %v\n", x.Value)
				}
			}
		}
		return true
	})
}

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetApproximateMatches(tree Tree, queries []string, distance int) *map[string][]*ResultInfo {
	results := map[string][]*Node{}
	matchIndex := map[string]bool{}

	for _, query := range queries {
		matches := tree.Search(query, distance)
		sort.Sort(ByNode(matches))
		matchHash := sha1.New()
		for _, match := range matches {
			io.WriteString(matchHash, fmt.Sprintf("%#v", match))
		}

		if len(matches) == 0 {
			continue
		}

		// Index queries with equal matches only once
		matchDigest := fmt.Sprintf("% x", matchHash.Sum(nil))
		if _, ok := matchIndex[matchDigest]; !ok {
			results[query] = matches
			matchIndex[matchDigest] = true
		}
	}

	/*
		Results are duplicated via their reciprocals:
			- query "hello"
				match in ".../example.go": "Hella"
			- query "hella"
				match in ".../example.go": "Hello"

		A last processing step over the results is necessary in order to group these
	*/
	groupedResults := groupResults(results)

	return groupedResults
}

type ResultInfo struct {
	Node              *Node
	AssociatedQueries string
}

type ResultCandidate struct {
	Query     string
	NodeWords string
	Result    []*Node
}

func groupResults(results map[string][]*Node) *map[string][]*ResultInfo {
	candidates := &[]*ResultCandidate{}

	for query, nodes := range results {
		wordAcc := &[]string{}
		for _, node := range nodes {
			*wordAcc = append(*wordAcc, node.Word)
		}
		sort.Strings(*wordAcc)
		*candidates = append(*candidates, &ResultCandidate{
			Query:     query,
			Result:    results[query],
			NodeWords: strings.ToLower(strings.Join(*wordAcc, "")),
		})
	}

	groupedResults := removeDuplicates(candidates)
	return groupedResults
}

func removeDuplicates(elements *[]*ResultCandidate) *map[string][]*ResultInfo {
	groupedResults := map[string][]*ResultInfo{}
	encountered := map[string]bool{}

	for _, v := range *elements {
		keys := []string{strings.ToLower(v.Query), v.NodeWords}
		sort.Strings(keys)
		compositeKey := strings.Join(keys, "")

		if encountered[compositeKey] == false {
			encountered[compositeKey] = true
			groupedResults[v.Query] = []*ResultInfo{}

			for _, result := range v.Result {
				groupedResults[v.Query] = append(groupedResults[v.Query], &ResultInfo{
					Node:              result,
					AssociatedQueries: v.Query + " <-> " + result.Word,
				})
			}
		}
	}

	return &groupedResults
}

func IndexSourcesFromPath(pathArgPtr *string) (Tree, []string) {
	fileList := getSourceFiles(*pathArgPtr)
	tree := Tree{}
	queries := []string{}
	for _, fileName := range fileList {
		inspect(fileName, &tree, &queries)
	}

	return tree, queries
}

func getSourceFiles(path string) []string {
	var files []string
	filepath.Walk(path, func(path string, f os.FileInfo, _ error) error {
		if f != nil && !f.IsDir() {
			if filepath.Ext(path) == ".go" {
				files = append(files, path)
			}
		}
		return nil
	})
	return files
}

func ParseArgs() (*int, string) {
	pathArgPtr := flag.String("dir", "", "Path containing the source to analyze. If none, will use os.Getwd()")
	distanceArgPtr := flag.Int("dist", 2, "Levenshtein-Damerau distance threshold. Default is 2")
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

	return distanceArgPtr, sourcePath
}
