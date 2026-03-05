package main

import (
	"reflect"
	"testing"
)

func TestMissingInTarget_BothEmpty(t *testing.T) {
	got := missingInTarget(nil, nil)
	if len(got) != 0 {
		t.Errorf("missingInTarget(nil, nil): want empty, got %v", got)
	}

	got = missingInTarget([]string{}, []string{})
	if len(got) != 0 {
		t.Errorf("missingInTarget([]string{}, []string{}): want empty, got %v", got)
	}
}

func TestMissingInTarget_SourceEmpty(t *testing.T) {
	got := missingInTarget([]string{}, []string{"a", "b"})
	if len(got) != 0 {
		t.Errorf("missingInTarget([], [a,b]): want empty, got %v", got)
	}
}

func TestMissingInTarget_TargetEmpty(t *testing.T) {
	source := []string{"c", "a", "b"}
	got := missingInTarget(source, []string{})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("missingInTarget(%v, []): want %v, got %v", source, want, got)
	}
}

func TestMissingInTarget_NoOverlap(t *testing.T) {
	source := []string{"x", "y", "z"}
	target := []string{"a", "b"}
	got := missingInTarget(source, target)
	want := []string{"x", "y", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("missingInTarget(%v, %v): want %v, got %v", source, target, want, got)
	}
}

func TestMissingInTarget_TargetHasAll(t *testing.T) {
	source := []string{"a", "b", "c"}
	target := []string{"a", "b", "c"}
	got := missingInTarget(source, target)
	if len(got) != 0 {
		t.Errorf("missingInTarget(%v, %v): want empty, got %v", source, target, got)
	}
}

func TestMissingInTarget_Mixed(t *testing.T) {
	source := []string{"a", "b", "c", "d"}
	target := []string{"a", "c"}
	got := missingInTarget(source, target)
	want := []string{"b", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("missingInTarget(%v, %v): want %v, got %v", source, target, want, got)
	}
}

func TestMissingInTarget_ResultSorted(t *testing.T) {
	source := []string{"z", "m", "a"}
	target := []string{}
	got := missingInTarget(source, target)
	want := []string{"a", "m", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("missingInTarget(%v, []): want sorted %v, got %v", source, want, got)
	}
}
