// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package annotations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gcla/termshark/v2/configs/profiles"
)

//======================================================================

type Annotation struct {
	PacketNum  int       `json:"packet_num"`
	Text       string    `json:"text"`
	ModifiedAt time.Time `json:"modified_at"`
}

//======================================================================

type AnnotationStore struct {
	mu       sync.Mutex
	entries  map[int]Annotation
	filePath string
}

//======================================================================

func NewStore(profileDir string) *AnnotationStore {
	name := profiles.CurrentName()
	dir := filepath.Join(profileDir, name)
	fp := filepath.Join(dir, ".termshark-annotations.json")
	s := &AnnotationStore{
		entries:  make(map[int]Annotation),
		filePath: fp,
	}
	_ = s.Load()
	return s
}

func (s *AnnotationStore) Get(packetNum int) (Annotation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.entries[packetNum]
	return a, ok
}

func (s *AnnotationStore) Set(packetNum int, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[packetNum] = Annotation{
		PacketNum:  packetNum,
		Text:       text,
		ModifiedAt: time.Now(),
	}
}

func (s *AnnotationStore) Delete(packetNum int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, packetNum)
}

func (s *AnnotationStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0666)
}

func (s *AnnotationStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &s.entries)
}

func (s *AnnotationStore) FilePath() string {
	return s.filePath
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
