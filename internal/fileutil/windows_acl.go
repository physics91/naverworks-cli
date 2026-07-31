package fileutil

import (
	"fmt"
	"strings"
)

// AllowedWindowsACLPrincipals are built-in principals that already have
// effective access to every file on the machine, so reporting them adds no
// protection. Every other principal must match the current user exactly.
var AllowedWindowsACLPrincipals = []string{
	`nt authority\system`,
	`builtin\administrators`,
}

// EvaluateWindowsACL parses `icacls <path>` output and returns a description of
// the first ACE whose principal is neither the current user nor a built-in
// administrative principal. An empty string means the ACL is acceptable.
//
// Principals are parsed and compared exactly (case-insensitive) rather than
// substring-tested, so entries such as `BUILTIN\Users` or `Everyone` cannot pass
// merely by containing the user name. Output with no recognizable ACE is
// reported as a problem (fail closed).
//
// This lives in fileutil so the private-key check and the credential-file check
// share one implementation, and so it stays testable on every platform.
func EvaluateWindowsACL(out, path, user string) string {
	sawACE := false
	for _, line := range strings.Split(out, "\n") {
		principal, ok := WindowsACLPrincipal(line, path)
		if !ok {
			continue
		}
		sawACE = true
		if IsAllowedWindowsACLPrincipal(principal, user) {
			continue
		}
		return fmt.Sprintf("%s에 현재 사용자(%s) 외의 접근 권한(%s)이 설정되어 있습니다", path, user, principal)
	}
	if !sawACE {
		return fmt.Sprintf("%s의 ACL 항목을 확인할 수 없습니다", path)
	}
	return ""
}

// WindowsACLPrincipal extracts the principal from a single icacls output line.
// The first line is prefixed with the inspected path, and each ACE line has the
// form `PRINCIPAL:(RIGHTS)`.
func WindowsACLPrincipal(line, path string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "Successfully") || strings.HasPrefix(trimmed, "성공") {
		return "", false
	}
	if strings.HasPrefix(trimmed, path) {
		trimmed = strings.TrimSpace(trimmed[len(path):])
	}
	idx := strings.Index(trimmed, ":(")
	if idx <= 0 {
		return "", false
	}
	principal := strings.TrimSpace(trimmed[:idx])
	if principal == "" {
		return "", false
	}
	return principal, true
}

// IsAllowedWindowsACLPrincipal reports whether an ACE principal is acceptable
// for the given identity.
//
// Comparison is exact (case-insensitive) on the fully qualified
// `DOMAIN\account` form. Matching on the bare account name was tried and
// removed: it let a principal from a different domain (`OTHERDOMAIN\bob`) be
// accepted for `bob`. Callers must therefore supply a domain-qualified identity;
// IsUsableWindowsIdentity enforces that.
func IsAllowedWindowsACLPrincipal(principal, user string) bool {
	p := strings.ToLower(strings.TrimSpace(principal))
	u := strings.ToLower(strings.TrimSpace(user))
	if p == "" || u == "" {
		return false
	}
	if p == u {
		return true
	}
	for _, allowed := range AllowedWindowsACLPrincipals {
		if p == allowed {
			return true
		}
	}
	return false
}

// ForeignWindowsACLPrincipals returns the distinct principals present in icacls
// output that are not permitted for the given user.
//
// `icacls /grant:r` only replaces entries for the principal it names, so an
// explicit ACE for another principal (for example a hand-added `Everyone`)
// survives an exclusive grant. Callers use this list to remove those entries.
func ForeignWindowsACLPrincipals(out, path, user string) []string {
	seen := make(map[string]bool)
	var foreign []string
	for _, line := range strings.Split(out, "\n") {
		principal, ok := WindowsACLPrincipal(line, path)
		if !ok || IsAllowedWindowsACLPrincipal(principal, user) {
			continue
		}
		key := strings.ToLower(principal)
		if seen[key] {
			continue
		}
		seen[key] = true
		foreign = append(foreign, principal)
	}
	return foreign
}

// disallowedIdentityNames must never be accepted as the current-user identity.
// The identity is the reference an ACL is compared against, so if it resolves to
// a broad or administrative principal, every matching ACE looks legitimate. This
// matters because the identity can fall back to the attacker-controllable
// USERNAME environment variable when `whoami` is unavailable.
var disallowedIdentityNames = []string{
	"everyone",
	"users",
	"authenticated users",
	"interactive",
	"network",
	"guest",
	"guests",
	"all application packages",
	"administrator",
	"administrators",
	"system",
	"trustedinstaller",
}

// IsUsableWindowsIdentity reports whether a resolved account name is safe to use
// as the reference identity in an ACL comparison.
//
// A domain (or machine) qualifier is required. ACL comparison is exact, so an
// unqualified name could never match a real `MACHINE\user` ACE anyway, and
// accepting one would previously have enabled cross-domain confusion.
func IsUsableWindowsIdentity(user string) bool {
	u := strings.ToLower(strings.TrimSpace(user))
	if u == "" {
		return false
	}

	idx := strings.LastIndex(u, `\`)
	if idx <= 0 || idx == len(u)-1 {
		return false
	}

	name := u[idx+1:]
	for _, bad := range disallowedIdentityNames {
		if u == bad || name == bad {
			return false
		}
	}
	// These are acceptable inside an ACL but must not stand in as the identity.
	for _, allowed := range AllowedWindowsACLPrincipals {
		if u == allowed {
			return false
		}
	}
	return true
}
