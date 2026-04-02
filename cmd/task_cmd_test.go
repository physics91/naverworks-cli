package cmd

import "testing"

func TestBuildTaskMoveBody_UsesToCategoryID(t *testing.T) {
	body := buildTaskMoveBody("cat1")

	if got := body["toCategoryId"]; got != "cat1" {
		t.Fatalf("toCategoryId = %v, want %q", got, "cat1")
	}
	if _, exists := body["taskCategoryId"]; exists {
		t.Fatal("taskCategoryId should not be present in move payload")
	}
}
