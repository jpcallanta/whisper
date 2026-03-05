package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/spf13/cobra"
)

var (
	sourceRegion string
	destRegion   string
	secretID     string
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate secrets from one AWS region to another",
	RunE:  runMigrate,
}

// Registers migrate flags and marks source/dest region required.
func init() {
	migrateCmd.Flags().StringVar(&sourceRegion, "source-region", "", "AWS region to read secrets from (required)")
	migrateCmd.Flags().StringVar(&destRegion, "dest-region", "", "AWS region to write secrets to (required)")
	migrateCmd.Flags().StringVar(&secretID, "secret-id", "", "Single secret name or ARN to migrate; if omitted, all secrets are migrated")

	_ = migrateCmd.MarkFlagRequired("source-region")
	_ = migrateCmd.MarkFlagRequired("dest-region")
}

// Matches AWS region codes (e.g. us-east-1, eu-west-1, us-gov-east-1).
var awsRegionPattern = regexp.MustCompile(`^[a-z]{2}-[a-z0-9-]+-\d+$`)

// Reports whether the string matches the AWS region naming format.
func isValidRegion(region string) bool {
	return awsRegionPattern.MatchString(region)
}

// Validates flags and runs the migration (single secret or all).
func runMigrate(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	// Require non-empty regions so we never call AWS with invalid config.
	if sourceRegion == "" || destRegion == "" {
		return fmt.Errorf("source-region and dest-region are required")
	}

	// Reject invalid region format so we show error and usage before calling AWS.
	if !isValidRegion(sourceRegion) {
		return fmt.Errorf("invalid source-region %q", sourceRegion)
	}
	if !isValidRegion(destRegion) {
		return fmt.Errorf("invalid dest-region %q", destRegion)
	}

	sourceClient, err := newSecretsManagerClient(ctx, sourceRegion)
	if err != nil {

		return fmt.Errorf("source region client: %w", err)
	}

	destClient, err := newSecretsManagerClient(ctx, destRegion)
	if err != nil {

		return fmt.Errorf("dest region client: %w", err)
	}

	// Single secret migration when --secret-id is set.
	if secretID != "" {
		return copySecret(ctx, sourceClient, destClient, secretID)
	}

	return copyAllSecrets(ctx, sourceClient, destClient)
}

// Builds a Secrets Manager client for the region using default credentials.
func newSecretsManagerClient(ctx context.Context, region string) (*secretsmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {

		return nil, err
	}

	return secretsmanager.NewFromConfig(cfg), nil
}

// Writes one secret to dest: CreateSecret, or PutSecretValue if it already exists.
func writeSecretToDest(ctx context.Context, dest *secretsmanager.Client, name string, secretString *string, secretBinary []byte) error {
	namePtr := &name
	createInput := &secretsmanager.CreateSecretInput{
		Name:         namePtr,
		SecretString: secretString,
		SecretBinary: secretBinary,
	}

	_, err := dest.CreateSecret(ctx, createInput)

	// Create failed; if resource exists, update via PutSecretValue instead.
	if err != nil {
		var existsErr *types.ResourceExistsException
		if !errors.As(err, &existsErr) {

			return fmt.Errorf("create secret %q: %w", name, err)
		}

		_, putErr := dest.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
			SecretId:     namePtr,
			SecretString: secretString,
			SecretBinary: secretBinary,
		})
		if putErr != nil {

			return fmt.Errorf("put secret %q: %w", name, putErr)
		}
		log.Printf("Updated existing secret %q in destination", name)

		return nil
	}

	log.Printf("Created secret %q in destination", name)

	return nil
}

// Copies one secret from source to dest (create or update).
func copySecret(ctx context.Context, source, dest *secretsmanager.Client, id string) error {
	out, err := source.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &id,
	})

	// GetSecretValue failed.
	if err != nil {

		return fmt.Errorf("get secret %q: %w", id, err)
	}

	return writeSecretToDest(ctx, dest, *out.Name, out.SecretString, out.SecretBinary)
}

const listSecretsPageSize = 100
const batchGetSecretValueSize = 20

// int32Ptr returns a pointer to v for API input fields that require *int32.
func int32Ptr(v int32) *int32 {
	return &v
}

// Fetches secret values for the given names in batches of batchGetSecretValueSize; returns all entries or an error.
func batchGetAllSecretValues(ctx context.Context, client *secretsmanager.Client, names []string) ([]types.SecretValueEntry, error) {
	var all []types.SecretValueEntry

	for i := 0; i < len(names); i += batchGetSecretValueSize {
		end := i + batchGetSecretValueSize
		if end > len(names) {
			end = len(names)
		}
		chunk := names[i:end]
		batchOut, err := client.BatchGetSecretValue(ctx, &secretsmanager.BatchGetSecretValueInput{
			SecretIdList: chunk,
		})

		if err != nil {

			return nil, fmt.Errorf("batch get secrets: %w", err)
		}

		if len(batchOut.Errors) > 0 {
			e := batchOut.Errors[0]
			id := ""
			if e.SecretId != nil {
				id = *e.SecretId
			}

			return nil, fmt.Errorf("get secret %q: %s: %s", id, ptrToStr(e.ErrorCode), ptrToStr(e.Message))
		}

		for _, entry := range batchOut.SecretValues {
			if entry.Name == nil {

				return nil, fmt.Errorf("batch response entry missing name")
			}
			all = append(all, entry)
		}
	}

	return all, nil
}

// Lists secrets in source and copies each to dest using batched reads.
func copyAllSecrets(ctx context.Context, source, dest *secretsmanager.Client) error {
	names, err := listSecretNames(ctx, source)
	if err != nil {

		return err
	}

	entries, err := batchGetAllSecretValues(ctx, source, names)
	if err != nil {

		return err
	}

	for _, entry := range entries {
		if err := writeSecretToDest(ctx, dest, *entry.Name, entry.SecretString, entry.SecretBinary); err != nil {

			return fmt.Errorf("copy %q: %w", *entry.Name, err)
		}
	}

	log.Printf("Migrated %d secret(s)", len(names))

	return nil
}

// ptrToStr returns the string pointed to by s, or empty string if nil.
func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
