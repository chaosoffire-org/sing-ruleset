package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sing-ruleset/internal/domain"
	"testing"
	"time"
)

func TestNewHTTPDownloader(t *testing.T) {
	client := &http.Client{}
	downloader := NewHTTPDownloader(client)

	if downloader == nil {
		t.Fatal("NewHTTPDownloader() should not return nil")
	}

	if downloader.client != client {
		t.Error("NewHTTPDownloader() client not set correctly")
	}
}

func TestHttpDownloader_Download_Success(t *testing.T) {
	// Create a test server
	expectedContent := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))

	defer server.Close()

	// Create a temporary directory
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "downloaded.txt")

	// Test download
	downloader := NewHTTPDownloader(server.Client())
	ctx := context.Background()

	err := downloader.Download(ctx, server.URL, filePath)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != expectedContent {
		t.Errorf("Download() content = %v, want %v", string(content), expectedContent)
	}
}

func TestHttpDownloader_Download_CreatesDirectory(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content"))
	}))
	defer server.Close()

	// Create a temporary directory with nested path
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nested", "deep", "downloaded.txt")

	// Test download (should create nested directories)
	downloader := NewHTTPDownloader(server.Client())
	ctx := context.Background()

	err := downloader.Download(ctx, server.URL, filePath)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Download() should create nested directories and file")
	}
}

func TestHttpDownloader_Download_InvalidURL_Empty(t *testing.T) {
	downloader := NewHTTPDownloader(&http.Client{})
	ctx := context.Background()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")

	err := downloader.Download(ctx, "", filePath)
	if err == nil {
		t.Error("Download() should return error for empty URL")
	}
}

func TestHttpDownloader_Download_InvalidURL_NoProtocol(t *testing.T) {
	downloader := NewHTTPDownloader(&http.Client{})
	ctx := context.Background()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")

	err := downloader.Download(ctx, "example.com/file.txt", filePath)
	if err == nil {
		t.Error("Download() should return error for URL without http/https protocol")
	}
}

func TestHttpDownloader_Download_HTTPError(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")

	downloader := NewHTTPDownloader(server.Client())
	ctx := context.Background()

	err := downloader.Download(ctx, server.URL, filePath)
	if err == nil {
		t.Error("Download() should return error for HTTP 404")
	}
}

func TestHttpDownloader_Download_HTTPError500(t *testing.T) {
	// Create a test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")

	downloader := NewHTTPDownloader(server.Client())
	ctx := context.Background()

	err := downloader.Download(ctx, server.URL, filePath)
	if err == nil {
		t.Error("Download() should return error for HTTP 500")
	}
}

func TestHttpDownloader_Download_ContextCanceled(t *testing.T) {
	// Create a slow test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")

	downloader := NewHTTPDownloader(server.Client())

	// Create a context that will be canceled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := downloader.Download(ctx, server.URL, filePath)
	if err == nil {
		t.Error("Download() should return error when context is canceled")
	}
}

func TestHttpDownloader_ImplementsInterface(t *testing.T) {
	// Compile-time check that HTTPDownloader implements domain.Downloader
	var _ domain.Downloader = (*HTTPDownloader)(nil)
}
