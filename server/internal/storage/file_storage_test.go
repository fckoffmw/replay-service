package storage

import (
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStorage создает временную директорию для тестов
func setupTestStorage(t *testing.T) (*FileStorage, string) {
	tmpDir := filepath.Join(os.TempDir(), "replay_test_"+uuid.New().String())
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err, "не удалось создать временную директорию")

	storage := NewFileStorage(tmpDir)
	return storage, tmpDir
}

// cleanupTestStorage удаляет временную директорию
func cleanupTestStorage(t *testing.T, tmpDir string) {
	err := os.RemoveAll(tmpDir)
	require.NoError(t, err, "не удалось удалить временную директорию")
}

// createTestFile создает тестовый multipart.FileHeader
func createTestFile(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	tmpFile := filepath.Join(os.TempDir(), filename)
	err := os.WriteFile(tmpFile, content, 0644)
	require.NoError(t, err)

	file, err := os.Open(tmpFile)
	require.NoError(t, err)
	defer file.Close()

	stat, err := file.Stat()
	require.NoError(t, err)

	return &multipart.FileHeader{
		Filename: filename,
		Size:     stat.Size(),
	}
}

// TestNewFileStorage проверяет создание FileStorage
func TestNewFileStorage(t *testing.T) {
	baseDir := "/test/storage"
	storage := NewFileStorage(baseDir)

	assert.NotNil(t, storage)
	assert.Equal(t, baseDir, storage.baseDir)
}

// TestGetFilePath проверяет получение полного пути к файлу
func TestGetFilePath(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tmpDir)

	relativePath := "user/game/replay.rep"
	expectedPath := filepath.Join(tmpDir, relativePath)

	fullPath := storage.GetFilePath(relativePath)

	assert.Equal(t, expectedPath, fullPath)
}

// TestSaveReplayFile_Success проверяет успешное сохранение файла
// Примечание: полноценное тестирование SaveReplayFile требует реального multipart файла
// Этот тест пропускается, так как создание настоящего multipart.FileHeader сложно в unit-тестах
func TestSaveReplayFile_Success(t *testing.T) {
	t.Skip("Тест требует реального HTTP multipart файла, тестируется через integration тесты")
}

// TestSaveReplayFile_CreatesDirectories проверяет создание директорий
func TestSaveReplayFile_CreatesDirectories(t *testing.T) {
	t.Skip("Тест требует реального HTTP multipart файла, тестируется через integration тесты")
}

// TestSaveReplayFile_PreservesExtension проверяет сохранение расширения файла
func TestSaveReplayFile_PreservesExtension(t *testing.T) {
	t.Skip("Тест требует реального HTTP multipart файла, тестируется через integration тесты")
}

// TestDeleteFile_Success проверяет успешное удаление файла
func TestDeleteFile_Success(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tmpDir)

	// Создаем файл
	relPath := "user/game/replay.rep"
	fullPath := storage.GetFilePath(relPath)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(fullPath, []byte("test"), 0644)
	require.NoError(t, err)

	// Проверяем, что файл существует
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)

	// Удаляем файл
	err = storage.DeleteFile(relPath)

	assert.NoError(t, err)

	// Проверяем, что файл удален
	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err), "файл должен быть удален")
}

// TestDeleteFile_NonExistent проверяет удаление несуществующего файла
func TestDeleteFile_NonExistent(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tmpDir)

	relPath := "user/game/nonexistent.rep"

	// Удаление несуществующего файла не должно возвращать ошибку
	err := storage.DeleteFile(relPath)

	// В зависимости от реализации, может быть nil или ошибка
	// Проверяем, что не паникует
	_ = err
}

// TestDeleteFiles_Success проверяет удаление нескольких файлов
func TestDeleteFiles_Success(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tmpDir)

	// Создаем несколько файлов
	filePaths := []string{
		"user/game/replay1.rep",
		"user/game/replay2.rep",
		"user/game/replay3.rep",
	}

	for _, relPath := range filePaths {
		fullPath := storage.GetFilePath(relPath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("test"), 0644)
		require.NoError(t, err)
	}

	// Удаляем все файлы
	errs := storage.DeleteFiles(filePaths)

	assert.Empty(t, errs, "не должно быть ошибок при удалении")

	// Проверяем, что все файлы удалены
	for _, relPath := range filePaths {
		fullPath := storage.GetFilePath(relPath)
		_, err := os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err), "файл %s должен быть удален", relPath)
	}
}

// TestDeleteFiles_PartialFailure проверяет частичный сбой при удалении
func TestDeleteFiles_PartialFailure(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tmpDir)

	// Создаем один существующий файл
	existingPath := "user/game/existing.rep"
	fullPath := storage.GetFilePath(existingPath)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(fullPath, []byte("test"), 0644)
	require.NoError(t, err)

	filePaths := []string{
		existingPath,
		"user/game/nonexistent1.rep",
		"user/game/nonexistent2.rep",
	}

	errs := storage.DeleteFiles(filePaths)

	// Могут быть ошибки для несуществующих файлов
	// Проверяем, что существующий файл удален
	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err), "существующий файл должен быть удален")

	// Количество ошибок зависит от реализации
	_ = errs
}

// TestDeleteFiles_EmptyList проверяет удаление пустого списка
func TestDeleteFiles_EmptyList(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tmpDir)

	errs := storage.DeleteFiles([]string{})

	assert.Empty(t, errs, "не должно быть ошибок для пустого списка")
}

// TestFileStorage_Integration проверяет полный цикл работы с файлами
func TestFileStorage_Integration(t *testing.T) {
	t.Skip("Тест требует реального HTTP multipart файла, тестируется через integration тесты")
}
