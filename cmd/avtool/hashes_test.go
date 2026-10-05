package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, dbPath string, args ...string) string {
	t.Helper()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"--db-path", dbPath}, args...))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute(%v): %v\noutput: %s", args, err, buf.String())
	}
	return buf.String()
}

const testHash64 = "2baf1f40105d9501fe319a8ec463fdf4325a2a5df445adf3f572f626253678c9"

func TestHashesAddListRm(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "avtool.db")

	runCLI(t, dbPath, "hashes", "add", testHash64, "TestVirus")

	out := runCLI(t, dbPath, "hashes", "list")
	if !strings.Contains(out, testHash64) || !strings.Contains(out, "TestVirus") {
		t.Fatalf("list output = %q", out)
	}

	runCLI(t, dbPath, "hashes", "rm", testHash64)

	out = runCLI(t, dbPath, "hashes", "list")
	if strings.Contains(out, testHash64) {
		t.Fatalf("expected %s removed, got %q", testHash64, out)
	}
}

func TestHashesAddRejectsMalformedHash(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "avtool.db")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--db-path", dbPath, "hashes", "add", "not-a-hash", "TestVirus"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for malformed hash, got nil")
	}

	rootCmd.SetArgs([]string{"--db-path", dbPath, "hashes", "add", "deadbeef", "TestVirus"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for too-short hash, got nil")
	}
}

func TestHashesRmByName(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "avtool.db")
	runCLI(t, dbPath, "hashes", "add", testHash64, "TestVirus")

	runCLI(t, dbPath, "hashes", "rm", "--name", "TestVirus")

	out := runCLI(t, dbPath, "hashes", "list")
	if strings.Contains(out, testHash64) {
		t.Fatalf("expected %s removed by name, got %q", testHash64, out)
	}
}

func TestHashesRmByNameAmbiguous(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "avtool.db")
	other := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	runCLI(t, dbPath, "hashes", "add", testHash64, "DupName")
	runCLI(t, dbPath, "hashes", "add", other, "DupName")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--db-path", dbPath, "hashes", "rm", "--name", "DupName"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for ambiguous name, got nil")
	} else if !strings.Contains(err.Error(), "matches 2 entries") {
		t.Fatalf("error = %v, want ambiguous-name message", err)
	}

	out := runCLI(t, dbPath, "hashes", "list")
	if !strings.Contains(out, testHash64) || !strings.Contains(out, other) {
		t.Fatalf("ambiguous rm must not delete either entry, got %q", out)
	}
}

func TestHashesRmRequiresHashOrName(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "avtool.db")
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--db-path", dbPath, "hashes", "rm"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected usage error, got nil")
	}
}
