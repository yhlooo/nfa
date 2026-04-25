package eula

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContent(t *testing.T) {
	content := Content()
	assert.NotEmpty(t, content, "EULA content should not be empty")
	assert.Contains(t, content, "End User License Agreement", "EULA should contain expected heading")
}

func TestSHA256(t *testing.T) {
	// SHA256 应该与直接计算的结果一致
	content := Content()
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	assert.Equal(t, expected, SHA256())
}

func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := writeToFile(tmpDir)
	require.NoError(t, err)

	// 验证文件内容
	written, err := os.ReadFile(filepath.Join(tmpDir, eulaFileName))
	require.NoError(t, err)
	assert.Equal(t, Content(), string(written))
}

func TestReadSignedSHA256_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, signatureFileName)

	_, err := readSignedSHA256(path)
	assert.True(t, os.IsNotExist(err))
}

func TestWriteAndReadSignedSHA256(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, signatureFileName)

	sha := SHA256()
	err := writeSignedSHA256(path, sha)
	require.NoError(t, err)

	// 验证文件存在且内容正确
	read, err := readSignedSHA256(path)
	require.NoError(t, err)
	assert.Equal(t, sha, read)
}

func TestSHA256_Deterministic(t *testing.T) {
	// 多次调用应返回相同结果
	first := SHA256()
	for i := 0; i < 10; i++ {
		assert.Equal(t, first, SHA256())
	}
}
