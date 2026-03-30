// verify-coverage-ledger.go — Baseline Coverage Ledger 검증 스크립트
//
// Usage:
//   go run scripts/verify-coverage-ledger.go docs/coverage-ledger-existing.md
//
// 기능:
//   - Markdown 테이블에서 endpoint 목록 파싱
//   - 중복 endpoint 검출
//   - 총계 출력
//   - 도메인별 집계

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type endpoint struct {
	Number int
	Domain string
	Method string
	Path   string
	File   string
	CLI    string
	Smoke  string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run %s <ledger.md>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "파일 열기 실패: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var endpoints []endpoint
	tableRowRegex := regexp.MustCompile(`^\|\s*(\d+)\s*\|`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !tableRowRegex.MatchString(line) {
			continue
		}

		cols := splitTableRow(line)
		if len(cols) < 7 {
			continue
		}

		num := 0
		fmt.Sscanf(strings.TrimSpace(cols[0]), "%d", &num)
		if num == 0 {
			continue
		}

		ep := endpoint{
			Number: num,
			Domain: strings.TrimSpace(cols[1]),
			Method: strings.TrimSpace(cols[2]),
			Path:   strings.TrimSpace(cols[3]),
			File:   strings.TrimSpace(cols[4]),
			CLI:    strings.TrimSpace(cols[5]),
			Smoke:  strings.TrimSpace(cols[6]),
		}
		endpoints = append(endpoints, ep)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "파일 읽기 오류: %v\n", err)
		os.Exit(1)
	}

	// --- 총계 ---
	fmt.Printf("=== Coverage Ledger 검증 결과 ===\n\n")
	fmt.Printf("총 endpoint 수: %d\n\n", len(endpoints))

	// --- 도메인별 집계 ---
	domainCount := make(map[string]int)
	domainOrder := []string{}
	seen := make(map[string]bool)
	for _, ep := range endpoints {
		domainCount[ep.Domain]++
		if !seen[ep.Domain] {
			seen[ep.Domain] = true
			domainOrder = append(domainOrder, ep.Domain)
		}
	}

	fmt.Printf("도메인별 분포:\n")
	for _, domain := range domainOrder {
		fmt.Printf("  %-25s %d\n", domain, domainCount[domain])
	}
	fmt.Println()

	// --- 중복 체크 ---
	type epKey struct {
		Method string
		Path   string
	}
	dupMap := make(map[epKey][]int)
	for _, ep := range endpoints {
		key := epKey{Method: ep.Method, Path: ep.Path}
		dupMap[key] = append(dupMap[key], ep.Number)
	}

	duplicates := []string{}
	for key, nums := range dupMap {
		if len(nums) > 1 {
			duplicates = append(duplicates, fmt.Sprintf("  %s %s → rows %v", key.Method, key.Path, nums))
		}
	}

	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		fmt.Printf("중복 발견: %d건\n", len(duplicates))
		for _, d := range duplicates {
			fmt.Println(d)
		}
	} else {
		fmt.Printf("중복: 없음\n")
	}
	fmt.Println()

	// --- 번호 연속성 체크 ---
	numberGaps := []int{}
	for i, ep := range endpoints {
		expected := i + 1
		if ep.Number != expected {
			numberGaps = append(numberGaps, expected)
		}
	}

	if len(numberGaps) > 0 {
		fmt.Printf("번호 불연속: %v\n", numberGaps)
	} else {
		fmt.Printf("번호 연속성: OK (1~%d)\n", len(endpoints))
	}
	fmt.Println()

	// --- Smoke test 현황 ---
	smokeCount := 0
	for _, ep := range endpoints {
		if ep.Smoke == "✓" {
			smokeCount++
		}
	}
	fmt.Printf("Smoke test 보유: %d / %d\n\n", smokeCount, len(endpoints))

	// --- 최종 판정 ---
	exitCode := 0
	if len(endpoints) != 116 {
		fmt.Printf("⚠  총계 불일치: 기대값 116, 실제 %d\n", len(endpoints))
		exitCode = 1
	} else {
		fmt.Printf("✓  총계 일치: 116개\n")
	}

	if len(duplicates) > 0 {
		exitCode = 1
	}

	if len(numberGaps) > 0 {
		exitCode = 1
	}

	os.Exit(exitCode)
}

// splitTableRow splits a Markdown table row by '|' and returns the cell contents.
// Leading and trailing empty strings from the split are removed.
func splitTableRow(line string) []string {
	parts := strings.Split(line, "|")
	// Remove first and last empty parts from leading/trailing '|'
	if len(parts) > 0 && strings.TrimSpace(parts[0]) == "" {
		parts = parts[1:]
	}
	if len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
