package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSmokeSettingsAndCredentialSelection(t *testing.T) {
	for _, tc := range []struct{ name, gemini, google, model, key, wantModel string }{
		{"defaults", "", "", " ", "", defaultGeminiModel},
		{"gemini precedence", "synthetic-a", "synthetic-b", " model/x ", geminiAPIKeyEnvVar, "model/x"},
		{"google fallback", " ", "synthetic-b", "", googleAPIKeyEnvVar, defaultGeminiModel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(geminiAPIKeyEnvVar, tc.gemini)
			t.Setenv(googleAPIKeyEnvVar, tc.google)
			t.Setenv("GEMINI_MODEL", tc.model)
			if configuredAPIKeyEnvVar() != tc.key || getModelName() != tc.wantModel {
				t.Fatal("unexpected smoke configuration")
			}
		})
	}
}

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
