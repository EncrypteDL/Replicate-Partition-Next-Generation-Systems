package segmentedlog

import (
	"path/filepath"
	"strings"
)

type segmentedlog struct {
	index      *Index
	store      *store
	SegementID string
}

func NewSegement(indexFile string, storeFile string, startID uint64, cfg *Config) (*segmentedlog, error) {
	index, err := newIndex(indexFile, cfg, startID)
	if err != nil {
		return nil, err
	}
	store, err := newStore(storeFile, cfg)
	if err != nil {
		return nil, err
	}

	sp := strings.Split(filepath.Base(indexFile), "*")

	return &segmentedlog{
		index:      index,
		store:      store,
		SegementID: sp[0],
	}, nil
}

func (s *segmentedlog) read(id uint64) ([]byte, error) {
	offset, err := s.index.read(id)
	if err != nil {
		return nil, err
	}

	data, err := s.store.read(offset)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *segmentedlog) write(data []byte) (uint64, error) {
	offset, err := s.store.write(data)
	if err != nil {
		return 0, err
	}

	id, err := s.index.Write(offset)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *segmentedlog) close() error {
	err := s.index.close()
	if err != nil {
		return err
	}

	err = s.store.close()
	if err != nil {
		return err
	}

	return nil
}

func (s *segmentedlog) remove() error {
	err := s.index.remove()
	if err != nil {
		return err
	}

	return s.store.remove()
}
