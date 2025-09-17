package main

import (
	"testing"
)

func TestParser(t *testing.T) {
	args := "echo 'aa bb'"

	ret := parseInput(args)

	if len(ret) == 0 {
		panic("not return val")
	}

	if ret[0] != "echo" || ret[1] != "aa bb" {
		t.Errorf("expected to parse it as two argument echo and 'aa bb', got: %v", ret)
	}

	args = "echo hello   world"

	ret = parseInput(args)

	if ret[0] != "echo" || ret[1] != "hello" || ret[2] != "world" {
		t.Errorf("expected to parse it as two argument echo hello world, got: %v", ret)
	}
}
