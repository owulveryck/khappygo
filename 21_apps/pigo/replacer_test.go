package main

import (
	"testing"
	"strings"
)
func TestReplacer(t *testing.T) {
	testImage := "projects/_/buckets/aerobic-botany-270918-input/objects/test4.jpg"
	expected := "gs://aerobic-botany-270918-input/test4.jpg"
	rplcr := strings.NewReplacer("projects/_/buckets/","gs://","objects/","")
	result := rplcr.Replace(testImage)
	if result != expected {
		t.Fatal(result)
	}
}