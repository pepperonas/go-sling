package storage

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type FileInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	IsDir      bool      `json:"isDir"`
	Path       string    `json:"path"`
	UploadedAt time.Time `json:"uploadedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Children   []string  `json:"children,omitempty"`
}

type Store struct {
	dataDir   string
	retention time.Duration
	maxSize   int64
	files     map[string]*FileInfo
	mu        sync.RWMutex
}

func New(dataDir string, retention time.Duration, maxSize int64) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	s := &Store{
		dataDir:   dataDir,
		retention: retention,
		maxSize:   maxSize,
		files:     make(map[string]*FileInfo),
	}

	s.loadMetadata()
	return s, nil
}

func (s *Store) DataDir() string {
	return s.dataDir
}

func (s *Store) List() []*FileInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files := make([]*FileInfo, 0, len(s.files))
	for _, f := range s.files {
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].UploadedAt.After(files[j].UploadedAt)
	})
	return files
}

func (s *Store) Get(id string) (*FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, ok := s.files[id]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", id)
	}
	return f, nil
}

func (s *Store) Save(name string, reader io.Reader, size int64) (*FileInfo, error) {
	if size > s.maxSize {
		return nil, fmt.Errorf("file too large: %d > %d", size, s.maxSize)
	}

	// Remove existing file with same name (overwrite)
	s.mu.RLock()
	for id, existing := range s.files {
		if existing.Name == name {
			s.mu.RUnlock()
			s.Delete(id)
			s.mu.RLock()
			break
		}
	}
	s.mu.RUnlock()

	id := generateID()
	dirPath := filepath.Join(s.dataDir, id)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dirPath, name)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	written, err := io.Copy(f, reader)
	if err != nil {
		os.RemoveAll(dirPath)
		return nil, err
	}

	info := &FileInfo{
		ID:         id,
		Name:       name,
		Size:       written,
		IsDir:      false,
		Path:       filePath,
		UploadedAt: time.Now(),
		ExpiresAt:  time.Now().Add(s.retention),
	}

	s.mu.Lock()
	s.files[id] = info
	s.mu.Unlock()

	s.saveMetadata()
	return info, nil
}

func (s *Store) SaveDirectory(baseName string, files map[string]io.Reader) (*FileInfo, error) {
	id := generateID()
	dirPath := filepath.Join(s.dataDir, id, baseName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}

	var totalSize int64
	var children []string

	for relPath, reader := range files {
		fullPath := filepath.Join(dirPath, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			os.RemoveAll(filepath.Join(s.dataDir, id))
			return nil, err
		}

		f, err := os.Create(fullPath)
		if err != nil {
			os.RemoveAll(filepath.Join(s.dataDir, id))
			return nil, err
		}

		n, err := io.Copy(f, reader)
		f.Close()
		if err != nil {
			os.RemoveAll(filepath.Join(s.dataDir, id))
			return nil, err
		}
		totalSize += n
		children = append(children, relPath)
	}

	info := &FileInfo{
		ID:         id,
		Name:       baseName,
		Size:       totalSize,
		IsDir:      true,
		Path:       dirPath,
		UploadedAt: time.Now(),
		ExpiresAt:  time.Now().Add(s.retention),
		Children:   children,
	}

	s.mu.Lock()
	s.files[id] = info
	s.mu.Unlock()

	s.saveMetadata()
	return info, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	info, ok := s.files[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("file not found: %s", id)
	}
	delete(s.files, id)
	s.mu.Unlock()

	// Remove the ID directory
	dirPath := filepath.Dir(info.Path)
	if info.IsDir {
		dirPath = filepath.Dir(dirPath)
	}
	os.RemoveAll(dirPath)
	s.saveMetadata()
	return nil
}

func (s *Store) StreamTarGz(w http.ResponseWriter, info *FileInfo) error {
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.tar.gz"`, info.Name))

	gw := gzip.NewWriter(w)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	basePath := info.Path
	return filepath.Walk(basePath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(filepath.Dir(basePath), path)

		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

func (s *Store) TotalSize() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var total int64
	for _, f := range s.files {
		total += f.Size
	}
	return total
}

func (s *Store) FileCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.files)
}

func (s *Store) metadataPath() string {
	return filepath.Join(s.dataDir, ".metadata.json")
}

func (s *Store) saveMetadata() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, _ := json.MarshalIndent(s.files, "", "  ")
	os.WriteFile(s.metadataPath(), data, 0644)
}

func (s *Store) loadMetadata() {
	data, err := os.ReadFile(s.metadataPath())
	if err != nil {
		return
	}
	json.Unmarshal(data, &s.files)

	// Clean up entries whose files no longer exist
	for id, info := range s.files {
		if _, err := os.Stat(info.Path); os.IsNotExist(err) {
			delete(s.files, id)
		}
	}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// SanitizePath ensures the path doesn't escape the data directory
func SanitizePath(base, path string) (string, error) {
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("invalid path: %s", path)
	}
	full := filepath.Join(base, cleaned)
	if !strings.HasPrefix(full, filepath.Clean(base)) {
		return "", fmt.Errorf("path escape attempt: %s", path)
	}
	return full, nil
}
