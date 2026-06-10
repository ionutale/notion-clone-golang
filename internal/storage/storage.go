package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidKey = errors.New("invalid storage key")

type FileStore interface {
	Put(ctx context.Context, key string, reader io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

type LocalFileStore struct {
	dir    string
	prefix string
}

func NewLocalFileStore(dir, urlPrefix string) *LocalFileStore {
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create storage dir %s: %v", dir, err))
	}
	return &LocalFileStore{dir: dir, prefix: urlPrefix}
}

func (s *LocalFileStore) validateKey(key string) error {
	if key == "" || strings.Contains(key, "..") || strings.Contains(key, "/") || strings.Contains(key, "\\") {
		return ErrInvalidKey
	}
	return nil
}

func (s *LocalFileStore) safePath(key string) (string, error) {
	if err := s.validateKey(key); err != nil {
		return "", err
	}
	return filepath.Join(s.dir, key), nil
}

func (s *LocalFileStore) Put(ctx context.Context, key string, reader io.Reader) error {
	path, err := s.safePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func (s *LocalFileStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := s.safePath(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *LocalFileStore) Delete(ctx context.Context, key string) error {
	path, err := s.safePath(key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (s *LocalFileStore) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", s.prefix, key)
}
