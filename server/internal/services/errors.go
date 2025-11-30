package services

import "fmt"

func wrapError(operation string, err error) error {
	return fmt.Errorf("failed to %s: %w", operation, err)
}

func notFoundError(entity string, err error) error {
	return fmt.Errorf("%s not found: %w", entity, err)
}
