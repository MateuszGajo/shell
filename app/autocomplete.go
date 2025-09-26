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

func findCommonPrefix(args []string) string {
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

func getMatchAutocompletion(autocompleteData []string, selected string) ([]string, []string) {
	var suffixAutocompletion []string
	autocompletion := []string{}
mainLoop:
	for _, item := range autocompleteData {
		i := 0
		for i = 0; i < len(selected) && len(item) > i; i++ {
			if selected[i] != item[i] {
				continue mainLoop
			}

		}
		if i == len(item) {
			continue
		}
		suffix := string(item[i:] + " ")
		suffixAutocompletion = append(suffixAutocompletion, suffix)
		autocompletion = append(autocompletion, item+" ")
	}
	return autocompletion, suffixAutocompletion
}

func getAutcompleteData() []string {
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

	return autocompleteData
}

func (a *AutoComplete) Do(line []rune, pos int) ([][]rune, int) {
	prefix := string(line[:pos])

	if a.lastPrefix == prefix {
		a.tabCount++
	} else {
		a.lastPrefix = prefix
		a.tabCount = 1
	}

	autoCompleteInput := getAutcompleteData()
	autoCompleteData, autoCompleteSuffixesData := getMatchAutocompletion(autoCompleteInput, prefix)
	commonPrefix := findCommonPrefix(autoCompleteData)
	if commonPrefix != "" && commonPrefix != prefix {
		return [][]rune{[]rune(commonPrefix[len(prefix):])}, pos + (len(commonPrefix) - len(prefix))
	}

	if len(autoCompleteSuffixesData) == 0 {
		return [][]rune{[]rune("\x07")}, 1
	}

	if len(autoCompleteSuffixesData) == 1 {
		return [][]rune{[]rune(autoCompleteSuffixesData[0])}, pos + len([]rune(autoCompleteSuffixesData[0]))
	}

	switch a.tabCount {
	case 0, 1:
		fmt.Fprint(os.Stderr, "\a")
		return nil, 1
	default:
		fmt.Println()
		fmt.Println(strings.Join(autoCompleteData, " "))
		fmt.Printf("$ %s", string(line))
		a.tabCount = 0
		return nil, 0
	}
}
