// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/gcla/termshark/v2/configs/profiles"
)

//======================================================================

type Preset struct {
	Name   string
	Filter string
}

type PresetStore struct {
	presets  []Preset
	filePath string
	mu       sync.Mutex
}

//======================================================================

func NewStore() (*PresetStore, error) {
	dir, err := profiles.CurrentDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	fp := filepath.Join(dir, ".termshark-presets.json")
	s := &PresetStore{
		presets:  make([]Preset, 0),
		filePath: fp,
	}
	_ = s.Load()
	return s, nil
}

func (s *PresetStore) List() []Preset {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := make([]Preset, len(s.presets))
	copy(res, s.presets)
	return res
}

func (s *PresetStore) Save(name string, filter string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.presets {
		if p.Name == name {
			s.presets[i].Filter = filter
			return s.Persist()
		}
	}
	s.presets = append(s.presets, Preset{Name: name, Filter: filter})
	return s.Persist()
}

func (s *PresetStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.presets {
		if p.Name == name {
			s.presets = append(s.presets[:i], s.presets[i+1:]...)
			return s.Persist()
		}
	}
	return nil
}

func (s *PresetStore) Get(name string) (Preset, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.presets {
		if p.Name == name {
			return p, true
		}
	}
	return Preset{}, false
}

func (s *PresetStore) Names() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := make([]string, len(s.presets))
	for i, p := range s.presets {
		res[i] = p.Name
	}
	return res
}

func (s *PresetStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.presets = make([]Preset, 0)
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.presets)
}

func (s *PresetStore) Persist() error {
	dir := filepath.Dir(s.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	data, err := json.Marshal(s.presets)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0666)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
