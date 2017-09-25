package typhoon

import (
	"crypto/sha1"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
)

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
				importStartPos = x.Pos()
				importEndPos = x.End()
			}
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				// Prevent duplicates
				if !inSlice(x.Value, *queries) {
					if x.Value == "" {
						return true
					}
					// Some import paths may still be here. We filter them by position
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

type ByNode []*Node

func (a ByNode) Len() int           { return len(a) }
func (a ByNode) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNode) Less(i, j int) bool { return a[i].Word < a[j].Word }

func GetApproximateMatches(tree Tree, queries []string, distance int) map[string][]*Node {
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

		matchDigest := fmt.Sprintf("% x", matchHash.Sum(nil))
		if _, ok := matchIndex[matchDigest]; !ok {
			results[query] = matches
			matchIndex[matchDigest] = true
		}
	}

	return results
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
