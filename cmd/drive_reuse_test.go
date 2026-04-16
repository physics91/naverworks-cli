package cmd

import (
	"testing"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

func TestFetchFilesWithOptionalFolderUsesRootValues(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("cursor", "", "")
	cmd.Flags().Int("count", 0, "")
	cmd.Flags().String("folder", "", "")
	if err := cmd.Flags().Set("cursor", "next"); err != nil {
		t.Fatalf("set cursor: %v", err)
	}
	if err := cmd.Flags().Set("count", "25"); err != nil {
		t.Fatalf("set count: %v", err)
	}

	var gotFolder, gotCursor string
	var gotCount int
	resp, err := fetchFilesWithOptionalFolder(cmd, func(folder, cursor string, count int) (*api.Response, error) {
		gotFolder = folder
		gotCursor = cursor
		gotCount = count
		return &api.Response{Body: []byte(`{}`)}, nil
	})
	if err != nil {
		t.Fatalf("fetchFilesWithOptionalFolder returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response, got nil")
	}
	if gotFolder != "" || gotCursor != "next" || gotCount != 25 {
		t.Fatalf("unexpected fetch args: folder=%q cursor=%q count=%d", gotFolder, gotCursor, gotCount)
	}
}

func TestFetchFilesWithOptionalFolderUsesFolderValues(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("cursor", "", "")
	cmd.Flags().Int("count", 0, "")
	cmd.Flags().String("folder", "", "")
	if err := cmd.Flags().Set("cursor", "next"); err != nil {
		t.Fatalf("set cursor: %v", err)
	}
	if err := cmd.Flags().Set("count", "25"); err != nil {
		t.Fatalf("set count: %v", err)
	}
	if err := cmd.Flags().Set("folder", "folder-1"); err != nil {
		t.Fatalf("set folder: %v", err)
	}

	var gotFolder, gotCursor string
	var gotCount int
	resp, err := fetchFilesWithOptionalFolder(cmd, func(folder, cursor string, count int) (*api.Response, error) {
		gotFolder = folder
		gotCursor = cursor
		gotCount = count
		return &api.Response{Body: []byte(`{}`)}, nil
	})
	if err != nil {
		t.Fatalf("fetchFilesWithOptionalFolder returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response, got nil")
	}
	if gotFolder != "folder-1" || gotCursor != "next" || gotCount != 25 {
		t.Fatalf("unexpected fetch args: folder=%q cursor=%q count=%d", gotFolder, gotCursor, gotCount)
	}
}
