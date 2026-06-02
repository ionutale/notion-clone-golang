package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
	os.MkdirAll(dir, 0755)
	return &LocalFileStore{dir: dir, prefix: urlPrefix}
}

func (s *LocalFileStore) Put(ctx context.Context, key string, reader io.Reader) error {
	path := filepath.Join(s.dir, key)
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func (s *LocalFileStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.dir, key))
}

func (s *LocalFileStore) Delete(ctx context.Context, key string) error {
	return os.Remove(filepath.Join(s.dir, key))
}

func (s *LocalFileStore) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", s.prefix, key)
}
