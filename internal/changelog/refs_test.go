package changelog

import (
	"reflect"
	"testing"
)

func TestExtractReferences_SubjectOnly(t *testing.T) {
	refs := ExtractReferences("Add login validation (#42)", "")
	want := []string{"#42"}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("got %v, want %v", refs, want)
	}
}

func TestExtractReferences_BodyOnly(t *testing.T) {
	refs := ExtractReferences("Update README", "Closes #15\nRefs #3")
	want := []string{"#15", "#3"}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("got %v, want %v", refs, want)
	}
}

func TestExtractReferences_Both(t *testing.T) {
	refs := ExtractReferences("Fix crash (#42)", "Fixes #15")
	want := []string{"#15", "#42"}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("got %v, want %v", refs, want)
	}
}

func TestExtractReferences_Deduplication(t *testing.T) {
	refs := ExtractReferences("Fix crash (#42)", "See also #42")
	want := []string{"#42"}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("got %v, want %v", refs, want)
	}
}

func TestExtractReferences_None(t *testing.T) {
	refs := ExtractReferences("Update README", "")
	if len(refs) != 0 {
		t.Errorf("got %v, want empty", refs)
	}
}
