package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractReleaseNoteSlugs(t *testing.T) {
	html := `
		<a href="https://naver.worksmobile.com/release-notes/core_20260409/">core</a>
		<a href="https://naver.worksmobile.com/release-notes/core_20260409/">dup</a>
		<script>
		const entries = [
			"https://naver.worksmobile.com/release-notes/drive_20260409/",
			"https://naver.worksmobile.com/release-notes/page/2/"
		];
		</script>
	`

	got := extractReleaseNoteSlugs(html)
	want := []string{"core_20260409", "drive_20260409"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected slugs: got %v want %v", got, want)
	}
}

func TestNormalizeReleaseNoteTitle(t *testing.T) {
	got := normalizeReleaseNoteTitle("웍스 드라이브 정기 업데이트 - 네이버웍스")
	if got != "웍스 드라이브 정기 업데이트" {
		t.Fatalf("unexpected title: %q", got)
	}
}

func TestBaselineDecisionsKeepIncompleteAPINotice(t *testing.T) {
	baselinePath := filepath.Join("..", "..", "docs", "baselines", "api-monitor-baseline.json")
	known, err := loadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	for _, slug := range []string{"core_20260514", "management_20260821"} {
		if _, ok := known[slug]; !ok {
			t.Errorf("impact-free notice %s must be baselined", slug)
		}
	}
	if _, ok := known["core_20260723"]; ok {
		t.Error("core_20260723 must remain tracked until issue #27 API work is complete")
	}
}
