package main

import "testing"

func TestNode_IsLowerThan(t *testing.T) {
	// Given
	node := Node{ID: "node-02"}
	expected := true

	// When
	actual := node.IsHigherThan("node-01")

	// Then
	if expected != actual {
		t.Fatalf("Actual %t is not equal to expected %t", actual, expected)
	}
}
