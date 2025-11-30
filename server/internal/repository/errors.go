package repository

import "fmt"

func wrapQueryError(operation string, err error) error {
	return fmt.Errorf("failed to %s: %w", operation, err)
}

func wrapScanError(entity string, err error) error {
	return fmt.Errorf("failed to scan %s: %w", entity, err)
}

func wrapNotFoundError(entity string) error {
	return fmt.Errorf("%s not found or access denied", entity)
}
