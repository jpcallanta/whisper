package main

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

func TestSecretEntryToCSVRow_AllNil(t *testing.T) {
	entry := types.SecretValueEntry{}
	got := secretEntryToCSVRow(entry)
	want := []string{"", "", "", ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("secretEntryToCSVRow(empty): want %v, got %v", want, got)
	}
}

func TestSecretEntryToCSVRow_NameAndARNOnly(t *testing.T) {
	name := "my-secret"
	arn := "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret"
	entry := types.SecretValueEntry{
		Name: &name,
		ARN:  &arn,
	}
	got := secretEntryToCSVRow(entry)
	want := []string{"my-secret", "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret", "", ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("secretEntryToCSVRow(name+arn): want %v, got %v", want, got)
	}
}

func TestSecretEntryToCSVRow_SecretStringWithSpecialChars(t *testing.T) {
	name := "s"
	secretStr := "line1\nline2,\"quoted\""
	entry := types.SecretValueEntry{
		Name:         &name,
		SecretString: &secretStr,
	}
	got := secretEntryToCSVRow(entry)
	want := []string{"s", "", "line1\nline2,\"quoted\"", ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("secretEntryToCSVRow(special chars): want %v, got %v", want, got)
	}
}

func TestSecretEntryToCSVRow_SecretBinary(t *testing.T) {
	name := "bin-secret"
	secretBin := []byte{0x00, 0xFF, 0x0a}
	entry := types.SecretValueEntry{
		Name:         &name,
		SecretBinary: secretBin,
	}
	got := secretEntryToCSVRow(entry)
	if len(got) != 4 {
		t.Fatalf("secretEntryToCSVRow(binary): want 4 columns, got %d", len(got))
	}
	if got[0] != "bin-secret" || got[1] != "" || got[2] != "" {
		t.Errorf("secretEntryToCSVRow(binary): name/arn/string columns want bin-secret, \"\", \"\"; got %q, %q, %q", got[0], got[1], got[2])
	}
	// 0x00 0xFF 0x0a in base64 is AP8K
	wantBase64 := "AP8K"
	if got[3] != wantBase64 {
		t.Errorf("secretEntryToCSVRow(binary): fourth column want base64 %q, got %q", wantBase64, got[3])
	}
}
