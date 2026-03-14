package util

import (
	"fmt"
	"testing"
)

func TestSplitSimpleCmd(t *testing.T) {
	res, err := SplitArgs("ls")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !listEquals([]string{"ls"}, res) {
		t.Fatalf("expected result does not match actual: %v", res)
	}
	fmt.Println(res)
}

func TestSplitCmdWithArgs(t *testing.T) {
	res, err := SplitArgs("ls -lhtr")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !listEquals([]string{"ls", "-lhtr"}, res) {
		t.Fatalf("expected result does not match actual: %v", res)
	}
	fmt.Println(res)
}

func TestSplitPipe(t *testing.T) {
	res, err := SplitArgs("ls -lhtr | wc -l")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !listEquals([]string{"ls", "-lhtr", "|", "wc", "-l"}, res) {
		t.Fatalf("expected result does not match actual: %v", res)
	}
	fmt.Println(res)
}

func TestSplitNestedInDouble(t *testing.T) {
	res, err := SplitArgs("kubectl get pods -o=jsonpath='{.items[?(@.metadata.labels.name==\"team\")].metadata.name}'")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !listEquals([]string{"kubectl", "get", "pods", "-o=jsonpath={.items[?(@.metadata.labels.name==\"team\")].metadata.name}"}, res) {
		t.Fatalf("expected result does not match actual: %#v", res)
	}
	fmt.Println(res)
}

func listEquals(a, b []string) bool {
	for i, ai := range a {
		if ai != b[i] {
			return false
		}
	}
	return true
}
