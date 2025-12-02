package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type FileStorage struct {
	baseDir string
}

func NewFileStorage(baseDir string) *FileStorage {
	return &FileStorage{
		baseDir: baseDir,
	}
}

func (fs *FileStorage) SaveReplayFile(file *multipart.FileHeader, userID, gameID, replayID uuid.UUID) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	relativePath := filepath.Join(userID.String(), gameID.String(), replayID.String()+filepath.Ext(file.Filename))
	fullPath := filepath.Join(fs.baseDir, relativePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return relativePath, nil
}

func (fs *FileStorage) DeleteFile(relativePath string) error {
	fullPath := filepath.Join(fs.baseDir, relativePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (fs *FileStorage) DeleteFiles(relativePaths []string) []error {
	var errors []error
	for _, path := range relativePaths {
		if err := fs.DeleteFile(path); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (fs *FileStorage) GetFilePath(relativePath string) string {
	return filepath.Join(fs.baseDir, relativePath)
}
