package web

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSessionProfileIsolation(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	first, err := SessionDir(configPath, "../member")
	if err != nil {
		t.Fatal(err)
	}
	second, err := SessionDir(configPath, "member")
	if err != nil {
		t.Fatal(err)
	}
	if first == second || filepath.Dir(first) != filepath.Join(filepath.Dir(configPath), "browser") {
		t.Fatalf("profiles escaped their isolated directory: %q, %q", first, second)
	}
	if err := prepareSession(first, "../member"); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		for _, path := range []string{filepath.Dir(first), first} {
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0700 {
				t.Fatalf("session directory is not owner-only: %s: %v", path, err)
			}
		}
	}
}

func TestSessionRejectsUnsafePaths(t *testing.T) {
	for _, dir := range []string{"", ".", t.TempDir()} {
		if err := prepareSession(dir, "default"); err == nil {
			t.Fatalf("accepted non-session directory: %q", dir)
		}
	}
	for _, profile := range []string{"", " "} {
		if _, err := SessionDir(filepath.Join(t.TempDir(), "config.json"), profile); err == nil {
			t.Fatal("accepted empty profile")
		}
	}
	if _, err := SessionDir("relative/config.json", "default"); err == nil {
		t.Fatal("accepted relative config path")
	}
}

func TestSessionRejectsSymlinks(t *testing.T) {
	root := t.TempDir()
	dir, err := SessionDir(filepath.Join(root, "config.json"), "default")
	if err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Dir(dir)); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	if err := prepareSession(dir, "default"); err == nil {
		t.Fatal("followed browser root symlink")
	}
	if entries, err := os.ReadDir(outside); err != nil || len(entries) != 0 {
		t.Fatalf("wrote through symlink: %v", err)
	}
}

func TestSessionRejectsProfileAndMetadataSymlinks(t *testing.T) {
	for _, targetKind := range []string{"profile", "metadata"} {
		t.Run(targetKind, func(t *testing.T) {
			dir, err := SessionDir(filepath.Join(t.TempDir(), "config.json"), "default")
			if err != nil {
				t.Fatal(err)
			}
			outside := t.TempDir()
			file := filepath.Join(outside, "preserve.json")
			if err := os.WriteFile(file, []byte("preserve"), 0600); err != nil {
				t.Fatal(err)
			}
			link, destination := dir, outside
			if targetKind == "metadata" {
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				link, destination = filepath.Join(dir, "owner.json"), file
			} else if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(destination, link); err != nil {
				if runtime.GOOS == "windows" {
					t.Skipf("symlink creation unavailable: %v", err)
				}
				t.Fatal(err)
			}
			if err := prepareSession(dir, "default"); err == nil {
				t.Fatal("followed session symlink")
			}
			if data, err := os.ReadFile(file); err != nil || string(data) != "preserve" {
				t.Fatalf("changed symlink destination: %v", err)
			}
		})
	}
}

func TestBrowserLogoutDeletesOnlyOwnedProfile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	first, _ := SessionDir(configPath, "default")
	second, _ := SessionDir(configPath, "other")
	for _, entry := range []struct{ dir, profile string }{{first, "default"}, {second, "other"}} {
		if err := prepareSession(entry.dir, entry.profile); err != nil {
			t.Fatal(err)
		}
	}
	tokenPath := filepath.Join(filepath.Dir(configPath), "token.json")
	if err := os.WriteFile(tokenPath, []byte("preserve-api-token"), 0600); err != nil {
		t.Fatal(err)
	}
	b := Browser{SessionDir: first, Profile: "default", ExecPath: "missing-browser"}
	if err := b.Logout(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatalf("local session remains: %v", err)
	}
	if _, err := os.Stat(filepath.Join(second, "owner.json")); err != nil {
		t.Fatalf("other profile was affected: %v", err)
	}
	if data, err := os.ReadFile(tokenPath); err != nil || string(data) != "preserve-api-token" {
		t.Fatalf("API token was affected: %v", err)
	}
	if err := b.Logout(); err != nil {
		t.Fatalf("repeated local logout failed: %v", err)
	}
}

func TestBrowserLogoutRejectsBusyAndUnownedSessions(t *testing.T) {
	for _, kind := range []string{"SingletonLock", "lockfile", "wrong-owner", "missing-owner"} {
		t.Run(kind, func(t *testing.T) {
			dir, _ := SessionDir(filepath.Join(t.TempDir(), "config.json"), "default")
			if err := prepareSession(dir, "default"); err != nil {
				t.Fatal(err)
			}
			path, content := filepath.Join(dir, kind), "busy"
			if kind == "missing-owner" {
				if err := os.Remove(filepath.Join(dir, "owner.json")); err != nil {
					t.Fatal(err)
				}
			} else {
				if kind == "wrong-owner" {
					path, content = filepath.Join(dir, "owner.json"), `{"kind":"browser-sessions","profile":"other"}`
				}
				if err := os.WriteFile(path, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := (Browser{SessionDir: dir, Profile: "default"}).Logout(); err == nil {
				t.Fatal("unsafe logout was accepted")
			}
			if _, err := os.Stat(dir); err != nil {
				t.Fatalf("unsafe session was deleted: %v", err)
			}
		})
	}
}

func TestBrowserLogoutRejectsSymlinksAndWrongProfilePath(t *testing.T) {
	for _, kind := range []string{"root", "profile", "metadata"} {
		t.Run(kind, func(t *testing.T) {
			dir, _ := SessionDir(filepath.Join(t.TempDir(), "config.json"), "default")
			outside := t.TempDir()
			preserve := filepath.Join(outside, "preserve.json")
			if err := os.WriteFile(preserve, []byte("preserve"), 0600); err != nil {
				t.Fatal(err)
			}
			link, destination := filepath.Dir(dir), outside
			if kind == "profile" {
				if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
					t.Fatal(err)
				}
				link = dir
			} else if kind == "metadata" {
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				link, destination = filepath.Join(dir, "owner.json"), preserve
			}
			if err := os.Symlink(destination, link); err != nil {
				if runtime.GOOS == "windows" {
					t.Skipf("symlink creation unavailable: %v", err)
				}
				t.Fatal(err)
			}
			if err := (Browser{SessionDir: dir, Profile: "default"}).Logout(); err == nil {
				t.Fatal("logout accepted a symlink")
			}
			if data, err := os.ReadFile(preserve); err != nil || string(data) != "preserve" {
				t.Fatalf("symlink target was affected: %v", err)
			}
		})
	}
	if err := (Browser{SessionDir: t.TempDir(), Profile: "default"}).Logout(); err == nil {
		t.Fatal("arbitrary path accepted for deletion")
	}
}
