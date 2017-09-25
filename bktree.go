/*
The MIT License (MIT)
Copyright (c) 2013 Taras Roshko, 2017 Luis Gomez

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package typhoon

import (
	"go/token"
	"strings"

	"github.com/antzucaro/matchr"
)

type Node struct {
	Word     string
	Children map[int]*Node
	Position *token.Position
}

type ByNode []*Node

func (a ByNode) Len() int           { return len(a) }
func (a ByNode) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNode) Less(i, j int) bool { return a[i].Word < a[j].Word }

func NewNode(word string, position *token.Position) *Node {
	return &Node{
		Word:     strings.ToLower(word),
		Children: nil,
		Position: position,
	}
}

func (n *Node) AddChild(key int, word string, position *token.Position) {
	if n.Children == nil {
		n.Children = make(map[int]*Node)
	}
	n.Children[key] = NewNode(word, position)
}

func (n *Node) Keys() []int {
	if n.Children == nil {
		return make([]int, 0)
	}
	var keys []int
	for key := range n.Children {
		keys = append(keys, key)
	}
	return keys
}

func (n *Node) Node(key int) *Node {
	if n.Children == nil {
		return nil
	}
	return n.Children[key]
}

func (n *Node) ContainsKey(key int) bool {
	if n.Children == nil {
		return false
	}
	_, ok := n.Children[key]
	return ok
}

type Tree struct {
	Root *Node
	Size int
}

func (tree *Tree) Add(word string, position *token.Position) {
	word = strings.ToLower(word)

	if tree.Root == nil {
		tree.Root = NewNode(word, position)
		tree.Size++
		return
	}

	curNode := tree.Root

	dist := matchr.DamerauLevenshtein(curNode.Word, word)
	for curNode.ContainsKey(dist) {
		curNode = curNode.Node(dist)
		dist = matchr.DamerauLevenshtein(curNode.Word, word)
	}

	// Exclude identical matches
	if dist != 0 {
		curNode.AddChild(dist, word, position)
		tree.Size++
	}
}

func (tree *Tree) Search(word string, distance int) []*Node {
	var matches = make([]*Node, 0, tree.Size)
	word = strings.ToLower(word)

	tree.RecursiveSearch(tree.Root, &matches, word, distance)

	return matches
}

func (tree *Tree) RecursiveSearch(node *Node, matches *[]*Node, word string, distance int) {
	curDist := matchr.DamerauLevenshtein(node.Word, word)
	minDist := curDist - distance
	maxDist := curDist + distance

	// Exclude identical matches
	if curDist != 0 && curDist <= distance {
		*matches = append(*matches, node)
	}

	for _, v := range node.Keys() {
		if minDist <= v && v <= maxDist {
			tree.RecursiveSearch(node.Node(v), matches, word, distance)
		}
	}
}
