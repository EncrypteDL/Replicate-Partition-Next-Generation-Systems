package segmentedlog


import (
	"encoding/binary"
	"errors"
	"os"

	"github.com/edsrzf/mmap-go"
)

// index will store mapping between recordID and recordOffset
// it will maintain it in memory and in index file
type Index struct {
	mm      mmap.MMap
	idxFile *os.File
	maxsize uint64
	size    uint64
	id      uint64
	startID uint64
}

var ErrMaxIndexSize = errors.New("max index size should be multiple by 16 and more than 0")
var ErrRecordNotFound   = errors.New("record is not found")

func (i *Index) Write(offset uint64) (uint64, error) {
	ii := (i.id - i.startID) * 16

	if ii >= i.maxsize {
		return 0, ErrMaxIndexSize
	}

	binary.BigEndian.PutUint64(i.mm[ii:ii+8], i.id)
	binary.BigEndian.PutUint64(i.mm[ii:ii+16], offset)

	i.size += 16
	i.id++

	return i.id - 1, i.mm.Flush()
}

func (i *Index) read(id uint64) (uint64, error) {
	ii := (id - i.startID) * 16
	if id == 0 || ii >= i.size {
		return 0, ErrRecordNotFound
	}

	sID := binary.BigEndian.Uint64(i.mm[ii : ii+8])
	sOffset := binary.BigEndian.Uint64(i.mm[ii+8 : ii+16])

	if sID != id {
		panic("write or read is not working correctly")
	}

	return sOffset, nil
}

func (i *Index) close() error {
	return i.idxFile.Close()
}

func (i *Index) remove() error {
	return os.Remove(i.idxFile.Name())
}

func newIndex(file string, cfg *Config, startID uint64) (*Index, error) {
	if startID == 0 {
		panic("recordID should not be zero")
	}

	if cfg.Segment.MaxIndexSizeBytes == 0 || cfg.Segment.MaxIndexSizeBytes%16 != 0 {
		return nil, ErrMaxIndexSize
	}

	_, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(file, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	err = os.Truncate(f.Name(), int64(cfg.Segment.MaxIndexSizeBytes))
	if err != nil {
		return nil, err
	}

	mm, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}

	var size uint64
	id := startID
	for i := 0; i < len(mm); i += 16 {
		b1 := binary.BigEndian.Uint64(mm[i : i+8])
		b2 := binary.BigEndian.Uint64(mm[i+8 : i+16])

		if b1 == 0 && b2 == 0 {
			break
		}

		size += 16
		id++
	}

	idx := &Index{
		mm:      mm,
		idxFile: f,
		maxsize: cfg.Segment.MaxIndexSizeBytes,
		size:    size,
		id:      id,
		startID: startID,
	}

	return idx, nil
}
