package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLookupCommitPRs_WithPR(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]pullResponse{
			{Number: 42, Body: "Implements feature X"},
		})
	}))
	defer srv.Close()

	c := &Client{Token: "test", Owner: "owner", Repo: "repo", BaseURL: srv.URL}
	refs, err := c.LookupCommitPRs("abc123")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0] != "#42" {
		t.Errorf("got %v, want [#42]", refs)
	}
}

func TestLookupCommitPRs_NoPR(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]pullResponse{})
	}))
	defer srv.Close()

	c := &Client{Token: "test", Owner: "owner", Repo: "repo", BaseURL: srv.URL}
	refs, err := c.LookupCommitPRs("abc123")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 0 {
		t.Errorf("got %v, want empty", refs)
	}
}

func TestLookupCommitPRs_WithLinkedIssues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]pullResponse{
			{Number: 42, Body: "Closes #15\nFixes #7"},
		})
	}))
	defer srv.Close()

	c := &Client{Token: "test", Owner: "owner", Repo: "repo", BaseURL: srv.URL}
	refs, err := c.LookupCommitPRs("abc123")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"#15", "#42", "#7"}
	if len(refs) != len(want) {
		t.Fatalf("got %v, want %v", refs, want)
	}
	for i, ref := range refs {
		if ref != want[i] {
			t.Errorf("ref[%d] = %s, want %s", i, ref, want[i])
		}
	}
}

func TestLookupCommitPRs_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := &Client{Token: "test", Owner: "owner", Repo: "repo", BaseURL: srv.URL}
	_, err := c.LookupCommitPRs("abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveCommitPRs_BatchWithErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/commits/bad/pulls" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode([]pullResponse{
			{Number: 10, Body: ""},
		})
	}))
	defer srv.Close()

	c := &Client{Token: "test", Owner: "owner", Repo: "repo", BaseURL: srv.URL}
	result := c.ResolveCommitPRs([]string{"good1", "bad", "good2"})

	if refs, ok := result["good1"]; !ok || len(refs) != 1 || refs[0] != "#10" {
		t.Errorf("good1: got %v", result["good1"])
	}
	if _, ok := result["bad"]; ok {
		t.Errorf("bad commit should not have results, got %v", result["bad"])
	}
	if refs, ok := result["good2"]; !ok || len(refs) != 1 || refs[0] != "#10" {
		t.Errorf("good2: got %v", result["good2"])
	}
}
