package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"
)

var (
	verifySourceRegion string
	verifyDestRegion   string
	verifyAddMissing   bool
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify secrets in target region against source and optionally add missing",
	RunE:  runVerify,
}

// Registers verify flags and marks source/dest region required.
func init() {
	verifyCmd.Flags().StringVar(&verifySourceRegion, "source-region", "", "AWS region to read secrets from (required)")
	verifyCmd.Flags().StringVar(&verifyDestRegion, "dest-region", "", "AWS region to compare and optionally write to (required)")
	verifyCmd.Flags().BoolVar(&verifyAddMissing, "add-missing", false, "Add secrets missing in target by copying from source")

	_ = verifyCmd.MarkFlagRequired("source-region")
	_ = verifyCmd.MarkFlagRequired("dest-region")
}

// Returns secret names present in source but not in target, sorted.
func missingInTarget(sourceNames, targetNames []string) []string {
	targetSet := make(map[string]struct{}, len(targetNames))
	for _, n := range targetNames {
		targetSet[n] = struct{}{}
	}

	var out []string
	for _, n := range sourceNames {
		if _, ok := targetSet[n]; !ok {
			out = append(out, n)
		}
	}

	sort.Strings(out)

	return out
}

// Returns all secret names for the client via paginated ListSecrets.
func listSecretNames(ctx context.Context, client *secretsmanager.Client) ([]string, error) {
	var names []string
	paginator := secretsmanager.NewListSecretsPaginator(client, &secretsmanager.ListSecretsInput{
		MaxResults: int32Ptr(listSecretsPageSize),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {

			return nil, fmt.Errorf("list secrets: %w", err)
		}

		for _, s := range page.SecretList {
			names = append(names, *s.Name)
		}
	}

	return names, nil
}

// Validates flags, compares source vs target secret names, reports missing, optionally adds them.
func runVerify(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	if verifySourceRegion == "" || verifyDestRegion == "" {
		return fmt.Errorf("source-region and dest-region are required")
	}

	if !isValidRegion(verifySourceRegion) {
		return fmt.Errorf("invalid source-region %q", verifySourceRegion)
	}
	if !isValidRegion(verifyDestRegion) {
		return fmt.Errorf("invalid dest-region %q", verifyDestRegion)
	}

	sourceClient, err := newSecretsManagerClient(ctx, verifySourceRegion)
	if err != nil {

		return fmt.Errorf("source region client: %w", err)
	}

	destClient, err := newSecretsManagerClient(ctx, verifyDestRegion)
	if err != nil {

		return fmt.Errorf("dest region client: %w", err)
	}

	sourceNames, err := listSecretNames(ctx, sourceClient)
	if err != nil {

		return fmt.Errorf("source: %w", err)
	}

	targetNames, err := listSecretNames(ctx, destClient)
	if err != nil {

		return fmt.Errorf("target: %w", err)
	}

	missing := missingInTarget(sourceNames, targetNames)

	if len(missing) == 0 {
		log.Printf("All %d source secret(s) present in target", len(sourceNames))

		return nil
	}

	log.Printf("Missing in target (%d): %v", len(missing), missing)

	if !verifyAddMissing {

		return nil
	}

	for _, name := range missing {
		if err := copySecret(ctx, sourceClient, destClient, name); err != nil {

			return fmt.Errorf("add missing %q: %w", name, err)
		}
	}

	log.Printf("Added %d missing secret(s) to target", len(missing))

	return nil
}
