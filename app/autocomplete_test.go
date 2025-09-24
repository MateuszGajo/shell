package main

import (
	"reflect"
	"testing"
)

func TestAutocomplete(t *testing.T) {
	autocomplete := &AutoComplete{}

	autocompletions, pos := autocomplete.Do([]rune("ech"), 3)

	if len(autocompletions) != 1 {
		t.Errorf("expected to found one autocompletion")
	}

	if !reflect.DeepEqual(autocompletions[0], []rune("o ")) {
		t.Error("Should auotcomplete \"o \"")
	}

	if pos != 5 {
		t.Errorf("should be at position 5")
	}

}
