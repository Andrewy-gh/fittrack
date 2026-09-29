package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLocalEnvRespectsPriorityAndSetenv(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ".env"), "GOOGLE_API_KEY=from-dotenv\nGEMINI_MODEL=env-model\n")
	writeTestFile(t, filepath.Join(dir, "setenv.sh"), "export GEMINI_API_KEY=from-setenv\nexport GEMINI_MODEL=setenv-model\n")

	chdirForTest(t, dir)
	unsetEnvForTest(t, geminiAPIKeyEnvVar, googleAPIKeyEnvVar, "GEMINI_MODEL")

	if err := loadLocalEnv(); err != nil {
		t.Fatalf("loadLocalEnv() error = %v", err)
	}

	if got := os.Getenv(geminiAPIKeyEnvVar); got != "from-setenv" {
		t.Fatalf("expected %s from setenv.sh, got %q", geminiAPIKeyEnvVar, got)
	}
	if got := os.Getenv(googleAPIKeyEnvVar); got != "from-dotenv" {
		t.Fatalf("expected %s from .env, got %q", googleAPIKeyEnvVar, got)
	}
	if got := os.Getenv("GEMINI_MODEL"); got != "env-model" {
		t.Fatalf("expected GEMINI_MODEL from .env, got %q", got)
	}
}

func unsetEnvForTest(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("os.Unsetenv(%q) error = %v", key, err)
		}

		t.Cleanup(func() {
			if !ok {
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("os.Unsetenv(%q) cleanup error = %v", key, err)
				}
				return
			}
			if err := os.Setenv(key, value); err != nil {
				t.Fatalf("os.Setenv(%q) cleanup error = %v", key, err)
			}
		})
	}
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd error = %v", err)
		}
	})
}

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}
