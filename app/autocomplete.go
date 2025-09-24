package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type AutoComplete struct {
	tabCount   int
	lastPrefix string
}

func commonPrefix(args []string, selected string) string {
	if len(args) < 2 {
		return ""
	}
	prefix := strings.TrimSpace(args[0])

	for _, item := range args[1:] {

		if !strings.Contains(item, prefix) {
			return ""
		}
	}

	return prefix
}

func (a *AutoComplete) Do(line []rune, pos int) ([][]rune, int) {
	selected := line[:pos]
	suffix := ""
	endPos := pos

	var suffixAutocompletion []string
	autocompletion := []string{}

	autocompleteData := []string{}

	for _, item := range builtinCommands {
		autocompleteData = append(autocompleteData, string(item))
	}

	files := displayFilesFromDir(os.Getenv("PATH"))

	autocompleteData = append(autocompleteData, files...)

	unique := make(map[string]bool)
	filtered := []string{}

	for _, item := range autocompleteData {
		if !unique[item] {
			unique[item] = true
			filtered = append(filtered, item)
		}
	}

	autocompleteData = filtered
	sort.Strings(autocompleteData)

mainLoop:
	for _, item := range autocompleteData {
		i := 0
		for i = 0; i < len(selected) && len(item) > i; i++ {
			if selected[i] != rune(item[i]) {
				continue mainLoop
			}

		}
		if i == len(item) {
			continue
		}
		suffix = string(item[i:] + " ")
		suffixAutocompletion = append(suffixAutocompletion, suffix)
		autocompletion = append(autocompletion, item+" ")
		endPos += len(suffix)
	}

	common := commonPrefix(suffixAutocompletion, string(selected))
	if common != "" {
		suffixAutocompletion = []string{common}
	}

	res := [][]rune{}

	for _, item := range suffixAutocompletion {
		res = append(res, []rune(item))
	}

	if len(res) == 0 {
		return [][]rune{[]rune("\x07")}, endPos
	}

	if len(res) > 1 {

		if a.lastPrefix == string(selected) {
			a.tabCount++
		} else {
			a.lastPrefix = string(selected)
			a.tabCount = 1
		}

		if a.tabCount == 1 {
			fmt.Fprint(os.Stderr, "\a")
			return nil, 1
		} else if a.tabCount == 2 {
			fmt.Println()

			fmt.Println(strings.Join(autocompletion, " "))
			fmt.Printf("$ %s", string(line))
			return nil, 0
		}
	}

	return res, endPos
}
