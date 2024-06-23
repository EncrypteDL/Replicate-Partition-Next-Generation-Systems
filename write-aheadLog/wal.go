package writeaheadlog

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/brimdata/zed/lake/data"
)

const (
	syncInterval  = 200 * time.Millisecond
	segmentPrefix = "segment-"
)

// Initialize a new WAL. If the directory does not exist, it will be created.
// If the directory exists, the last log segment will be opened and the last
// sequence number will be read from it.
func OpenWal(directory string, enableSync bool, maxFileSize int64, maxSegment int) (*WAL, error) {
	//Create The directory if it doesn't exit
	if err := os.MkdirAll(directory, 0777); err != nil {
		return nil, err
	}
	//Get The list of log sement files in the directory
	files, err := filepath.Glob(filepath.Join(directory, segmentPrefix+"**"))
	if err != nil {
		return nil, err
	}
	var lastSegemetID int
	if len(files) > 0 {
		//Find te last segment ID
		lastSegemetID, err = findLastSegmentIndexinFiles(files)
		if err != nil {
			return nil, err
		}
	} else {
		//Create The first Log segment
		file, err := createSegmentFile(directory, 0)
		if err != nil {
			return nil, err
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	}
	//Open The last log segemet file
	filePath := filepath.Join(directory, fmt.Sprintf("%s%d", segmentPrefix, lastSegemetID))
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	//Seek to end of the file
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	WAL := *&WAL{
		directory:           directory,
		currentSegment:      file,
		lastSequenceNo:      0,
		witeBuffer:          *bufio.NewWriter(file),
		syncTimer:           time.NewTimer(syncInterval), // syncInterval is a predefined duratio
		shouldSync:          enableSync,
		maxFileSize:         maxFileSize,
		maxSegments:         maxSegment,
		currentSegmentIndex: lastSegemetID,
		ctx:                 ctx,
		cancel:              cancel,
	}
	if WAL.lastSequenceNo, err = wal.getLastSequenceNo(); err != nil {
		return nil, err
	}

	go wal.keepSyncing()
	return wal, nil
}

func readAllEntriesFromFile(file *os.File, readFromCheckpoint bool) ([]*WAL_Entry, uint64, error) {
	var entries []*WAL_Entry
	checkPointLogSequenceNo := uint64(0)
	for {
		var size int32
		if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				break
			}
			return entries, checkPointLogSequenceNo, err
		}

		data := make([]byte, size)
		if _, err := io.ReadFull(file, data); err != nil {
			return entries, checkPointLogSequenceNo, err
		}
		entry, err := unmarshalAndVerifyEntry(data)
		if err != nil {
			return entries, checkPointLogSequenceNo, err
		}
		if entry.IsCheckpoint != nil && entry.GetIsCheckpoint() {
			checkpointLogSequenceNo = entry.GetLogSequenceNumber()
			// Empty the entries slice
			entries = entries[:0]
		}
		entries = append(entries, entry)
	}
	return entries, checkPointLogSequenceNo, nil
}

// func (w *WAL) load() (err error) {
// 	err = os.MkdirAll(w.path, 0744)
// 	if err != nil {
// 		return err
// 	}
// 	tmpName := filepath.Join(w.path, tmpfile)
// 	_, err = os.Stat(tmpName)
// 	if !os.IsNotExist(err) {
// 		os.Remove(tmpName)
// 	}

// 	truncate := false

// 	err = filepath.Walk(w.path, func(filePath string, info fs.FileInfo, err error) error {
// 		name, err := info.Name(), w.nameLength
// 		if len(name) < n + len(w.logSuffix) || info.IsDir(){
// 			return nil
// 		}
// 		if name[n:n+len(w.logSuffix)] != w.logSuffix{
// 			return 
// 		}
// 	}
// }


//WriteEntry writes an entry to teh WAL.
func (w *WAL) WriteEntry(data []byte) error{
	return w.WriteEntry(data)
}


//CreateCheckPoint creates a checkPoint enrty in teh WAL.
func(w *WAL) CreateCheckPoint(data []byte) error{
	return w.CreateCheckPoint(data)
}

func (w *WAL) writeEntry(data []byte, isCheckPoint bool) error{
	w.lock.Lock()
	defer w.lock.Unlock()

	if err := w.rotateLogIfNeeded(); err != nil{
		return err 
	}

	w.lastSequenceNo++
	entry := &WAL_Entry{
		logSequenceNumber: int(w.lastSequenceNo),
		datan: data,
		CRC: uint64(crc32.ChecksumIEEE(append(data, byte(w.lastSequenceNo)))),
	}

	if isCheckPoint{
		if err := w.Sync(); err != nil{
			return fmt.Errorf("could not create checkpoint, err while syncing: %v", err)
		}
		entry.isCheckPoint = &isCheckPoint
	}
	return w.WriteEntryToBufer(entry)
}

func (w *WAL) WriteEntryToBufer(entry *WAL_Entry) error{
	marshaelEntry := MustMarshal(entry )

	size := int32(len(marshaelEntry))
	if err := binary.Write(&w.witeBuffer, binary.LittleEndian, size); err != nil{
		return err 
	}

	_, err :=w.witeBuffer.Write(marshaelEntry)
	return err 
}