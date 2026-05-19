package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type MemoryStorage interface {
	LoadSession(ctx context.Context, sid string) (*SessionState, error)
	SaveSession(ctx context.Context, st *SessionState) error
	AppendTurn(ctx context.Context, sid string, turn Turn) error
}

type JSONLStorage struct {
	dir string
	mu  sync.Mutex
}

func NewJSONLStorage(dir string) *JSONLStorage {
	os.MkdirAll(dir, 0755)
	return &JSONLStorage{dir: dir}
}

func (s *JSONLStorage) LoadSession(ctx context.Context, sid string) (*SessionState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.filePath(sid)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return &SessionState{SessionID: sid, Turns: []Turn{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var turns []Turn
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var t Turn
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			continue
		}
		turns = append(turns, t)
	}

	return &SessionState{SessionID: sid, Turns: turns}, nil
}

func (s *JSONLStorage) SaveSession(ctx context.Context, st *SessionState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.filePath(st.SessionID)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	for _, t := range st.Turns {
		data, _ := json.Marshal(t)
		f.Write(data)
		f.WriteString("\n")
	}
	return nil
}

func (s *JSONLStorage) AppendTurn(ctx context.Context, sid string, turn Turn) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.filePath(sid)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	data, _ := json.Marshal(turn)
	f.Write(data)
	f.WriteString("\n")
	return nil
}

func (s *JSONLStorage) filePath(sid string) string {
	return filepath.Join(s.dir, sid+".jsonl")
}
