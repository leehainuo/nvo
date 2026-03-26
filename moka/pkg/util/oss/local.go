package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalConfig struct {
    Path    string `mapstructure:"path"`
    BaseURL string `mapstructure:"base_url"`
}

type localStorage struct {
	path     string
	baseURL  string
}

func initLocal(c *Config) (Storager, error) {
	if c.Local.Path == "" {
		return nil, fmt.Errorf("local path is required")
	}

	if err := os.MkdirAll(c.Local.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	local := &localStorage{
		path:    c.Local.Path,
		baseURL: c.Local.BaseURL,
	}

	return local, nil
}

func (l *localStorage) Upload(c context.Context, key string, reader io.Reader) (string, error) {
    filePath := filepath.Join(l.path, key)
    
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return "", fmt.Errorf("failed to create directory: %w", err)
    }
 
    file, err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()
 
    if _, err := io.Copy(file, reader); err != nil {
        return "", fmt.Errorf("failed to write file: %w", err)
    }
 
    return l.URL(c, key)
}

 
func (l *localStorage) Download(c context.Context, key string) (io.ReadCloser, error) {
    filePath := filepath.Join(l.path, key)
    
    file, err := os.Open(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("file not found: %s", key)
        }
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
 
    return file, nil
}

func (l *localStorage) Exists(c context.Context, key string) (bool, error) {
    filePath := filepath.Join(l.path, key)
    
    _, err := os.Stat(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, fmt.Errorf("failed to check file: %w", err)
    }
 
    return true, nil
}

func (l *localStorage) URL(c context.Context, key string) (string, error) {
    if l.baseURL == "" {
        return filepath.Join(l.path, key), nil
    }
    return fmt.Sprintf("%s/%s", l.baseURL, key), nil
}

func (l *localStorage) Delete(c context.Context, key string) error {
    filePath := filepath.Join(l.path, key)
    
    if err := os.Remove(filePath); err != nil {
        if os.IsNotExist(err) {
            return nil
        }
        return fmt.Errorf("failed to delete file: %w", err)
    }
 
    return nil
}

func (l *localStorage) Close() error {
    return nil
}