// Package service provides infrastructure service implementations.
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sing-ruleset/internal/domain"
	"strings"
)

// HTTPDownloader implements domain.Downloader using HTTP client.
type HTTPDownloader struct {
	client *http.Client
}

var _ domain.Downloader = (*HTTPDownloader)(nil)

// NewHTTPDownloader creates a new HTTPDownloader with the given HTTP client.
func NewHTTPDownloader(client *http.Client) *HTTPDownloader {
	return &HTTPDownloader{
		client: client,
	}
}

// Download downloads a file from the given URL and saves it to the specified file path.
func (d *HTTPDownloader) Download(ctx context.Context, url string, filePath string) error {
	errchan := make(chan error)

	go func() {
		defer close(errchan)

		if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
			errchan <- fmt.Errorf("url is not valid, url: %s", url)
			return
		}

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			errchan <- fmt.Errorf("failed to create directory: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			errchan <- err
			return
		}

		resp, err := d.client.Do(req)
		if err != nil {
			errchan <- err
			return
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				errchan <- fmt.Errorf("failed to close response body: %w", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			errchan <- fmt.Errorf("HTTP error: %s", resp.Status)
			return
		}

		out, err := os.Create(filePath)
		if err != nil {
			errchan <- err
			return
		}

		defer func() {
			if err := out.Close(); err != nil {
				errchan <- fmt.Errorf("failed to close output file: %w", err)
			}
		}()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			errchan <- err

			if closeErr := out.Close(); closeErr != nil {
				errchan <- fmt.Errorf("failed to close output file: %w", closeErr)
				return
			}

			if removeErr := os.Remove(filePath); removeErr != nil {
				errchan <- fmt.Errorf("failed to remove file: %w", removeErr)
				return
			}

			return
		}

		errchan <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errchan:
		return err
	}
}
