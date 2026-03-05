package main

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/spf13/cobra"
)

var (
	dumpRegion string
	dumpOutput string
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all secrets from a region to a CSV file",
	RunE:  runDump,
}

// Registers dump flags and marks region required.
func init() {
	dumpCmd.Flags().StringVar(&dumpRegion, "region", "", "AWS region to list and read secrets from (required)")
	dumpCmd.Flags().StringVarP(&dumpOutput, "output", "o", "", "Output file path; omit to write to stdout")

	_ = dumpCmd.MarkFlagRequired("region")
}

// Converts a SecretValueEntry into a CSV row: Name, ARN, SecretString, SecretBinary (base64).
func secretEntryToCSVRow(entry types.SecretValueEntry) []string {
	name := ptrToStr(entry.Name)
	arn := ptrToStr(entry.ARN)
	secretStr := ""
	if entry.SecretString != nil {
		secretStr = *entry.SecretString
	}
	secretBin := ""
	if len(entry.SecretBinary) > 0 {
		secretBin = base64.StdEncoding.EncodeToString(entry.SecretBinary)
	}

	return []string{name, arn, secretStr, secretBin}
}

// Validates flags, fetches all secrets in the region, and writes them to CSV.
func runDump(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	if dumpRegion == "" {
		return fmt.Errorf("region is required")
	}

	if !isValidRegion(dumpRegion) {
		return fmt.Errorf("invalid region %q", dumpRegion)
	}

	client, err := newSecretsManagerClient(ctx, dumpRegion)
	if err != nil {

		return fmt.Errorf("secrets manager client: %w", err)
	}

	names, err := listSecretNames(ctx, client)
	if err != nil {

		return fmt.Errorf("list secrets: %w", err)
	}

	var out io.Writer = os.Stdout
	if dumpOutput != "" {
		f, err := os.Create(dumpOutput)
		if err != nil {

			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	w := csv.NewWriter(out)
	defer w.Flush()

	header := []string{"Name", "ARN", "SecretString", "SecretBinary"}
	if err := w.Write(header); err != nil {

		return fmt.Errorf("write CSV header: %w", err)
	}

	if len(names) == 0 {

		return nil
	}

	entries, err := batchGetAllSecretValues(ctx, client, names)
	if err != nil {

		return fmt.Errorf("get secret values: %w", err)
	}

	for _, entry := range entries {
		row := secretEntryToCSVRow(entry)
		if err := w.Write(row); err != nil {

			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	return nil
}
