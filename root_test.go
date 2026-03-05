package main

import (
	"bytes"
	"strings"
	"testing"
)

// Ensures missing required flags cause help to be printed.
func TestExecute_MigrateMissingRequiredFlags_OutputsHelp(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	}()

	rootCmd.SetArgs([]string{"migrate", "--source-region", "", "--dest-region", ""})

	_ = Execute()

	help := out.String()
	if !strings.Contains(help, "Usage:") {
		t.Errorf("expected help output to contain Usage: got %q", help)
	}
	if !strings.Contains(help, "source-region") || !strings.Contains(help, "dest-region") {
		t.Errorf("expected help output to contain required flags: got %q", help)
	}
}

// Ensures invalid region strings cause error and usage to be printed.
func TestExecute_MigrateInvalidRegion_OutputsHelp(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	}()

	rootCmd.SetArgs([]string{"migrate", "--source-region", "not-a-region", "--dest-region", "us-east-1"})

	err := Execute()
	if err == nil {
		t.Fatal("Execute() with invalid source-region: expected error, got nil")
	}

	help := out.String()
	if !strings.Contains(help, "Usage:") {
		t.Errorf("expected help output to contain Usage: got %q", help)
	}
	if !strings.Contains(help, "source-region") || !strings.Contains(help, "dest-region") {
		t.Errorf("expected help output to contain required flags: got %q", help)
	}
}

// Ensures migrate --help runs without error and does not call AWS.
func TestExecute_MigrateHelp_Succeeds(t *testing.T) {
	rootCmd.SetArgs([]string{"migrate", "--help"})

	err := Execute()
	if err != nil {
		t.Fatalf("Execute() with migrate --help: %v", err)
	}
}

// Ensures root --help runs without error.
func TestExecute_RootHelp_Succeeds(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})

	err := Execute()
	if err != nil {
		t.Fatalf("Execute() with --help: %v", err)
	}
}
