package writeaheadlog

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	if WAL.lastSequenceNo, err = w.getLastSequenceNo(); err != nil {
		return nil, err
	}

	go w.keepSyncing()
	return w, nil
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
		// If we are reading from checkpoint and we find a checkpoint entry, we
		// we should return the entries from the last checkpoint. So we empty the
		// entries slice and start appending entries from the checkpoint.
		if entry.isCheckPoint != nil && entry.GetIsChechPoint() {
			checkPointLogSequenceNo = entry.GetLogSequenceNumber()
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

// WriteEntry writes an entry to teh WAL.
func (w *WAL) WriteEntry(data []byte) error {
	return w.WriteEntry(data)
}

// CreateCheckPoint creates a checkPoint enrty in teh WAL.
func (w *WAL) CreateCheckPoint(data []byte) error {
	return w.CreateCheckPoint(data)
}

func (w *WAL) writeEntry(data []byte, isCheckPoint bool) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if err := w.rotateLOgIFNedded(); err != nil {
		return err
	}

	w.lastSequenceNo++
	entry := &WAL_Entry{
		logSequenceNumber: int(w.lastSequenceNo),
		datan:             data,
		CRC:               uint64(crc32.ChecksumIEEE(append(data, byte(w.lastSequenceNo)))),
	}

	if isCheckPoint {
		if err := w.Sync(); err != nil {
			return fmt.Errorf("could not create checkpoint, err while syncing: %v", err)
		}
		entry.isCheckPoint = isCheckPoint
	}
	return w.WriteEntryToBufer(entry)
}

func (w *WAL) WriteEntryToBufer(entry *WAL_Entry) error {
	marshaelEntry := MustMarshal(entry)

	size := int32(len(marshaelEntry))
	if err := binary.Write(&w.witeBuffer, binary.LittleEndian, size); err != nil {
		return err
	}

	_, err := w.witeBuffer.Write(marshaelEntry)
	return err
}

func (w *WAL) rotateLOgIFNedded() error {
	fileInfo, err := w.currentSegment.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size()+int64(w.witeBuffer.Buffered()) >= w.maxFileSize {
		if err := w.rotateLog(); err != nil {
			return err
		}
	}
	return nil
}

func (w *WAL) rotateLog() error {
	if err := w.Sync(); err != nil {
		return err
	}

	if err := w.currentSegment.Close(); err != nil {
		return err
	}
	w.currentSegmentIndex++
	if w.currentSegmentIndex >= w.maxSegments {
		if err := w.deleteOldeseSegment(); err != nil {
			return err
		}
	}

	newFile, err := createSegmentFile(w.directory, w.currentSegmentIndex)
	if err != nil {
		return err
	}

	w.currentSegment = newFile
	w.witeBuffer = *bufio.NewWriter(newFile)

	return nil
}

// remove the oldest log file
func (w *WAL) deleteOldeseSegment() error {
	files, err := filepath.Glob(filepath.Join(w.directory, segmentPrefix+"**"))
	if err != nil {
		return err
	}

	var oldestSegmetFilePath string
	if len(files) > 0 {
		//Find the oldest segement ID
		oldestSegmetFilePath, err = w.findOldestSegemetFile(files)
		if err != nil {
			return err
		}
	} else {
		return nil
	}

	//Delete the oldes segemnt file
	if err = os.Remove(oldestSegmetFilePath); err != nil {
		return err
	}
	return nil
}

func (w *WAL) findOldestSegemetFile(files []string) (string, error) {
	var oldestSegmetFilePath string
	oldestSegemetID := math.MaxInt64
	for _, file := range files {
		//Get the segment index from the file name
		segemntIndex, err := strconv.Atoi(strings.TrimPrefix(file, filepath.Join(w.directory, segmentPrefix)))
		if err != nil {
			return "", err
		}

		if segemntIndex < oldestSegemetID {
			oldestSegemetID = segemntIndex
			oldestSegmetFilePath = file
		}
	}
	return oldestSegmetFilePath, nil
}

// Close Th wal file. It also calls Sync() on the Wal()
func (w *WAL) Close() error {
	w.cancel()
	if err := w.Sync(); err != nil {
		return err
	}
	// if err != w.Sync(); err != nil{
	// 	return err
	// }
	return w.currentSegment.Close()
}

// Read all entries from the WAL. If readFromCheckpoint is true, it will return
// all the entries from the last checkpoint (if no checkpoint is found, it will
// return an empty slice.)

func (w *WAL) ReadAll(readFromCheckpoint bool) ([]*WAL_Entry, error) {
	file, err := os.OpenFile(w.currentSegment.Name(), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries, checkpoint, err := readAllEntriesFromFile(file, readFromCheckpoint)
	if err != nil {
		return entries, err
	}
	if readFromCheckpoint && checkpoint <= 0 {
		return entries[:0], nil
	}

	return entries, nil
}

// Starts reading from log segment files starting from the given offset
// (Segment Index) and returns all the entries. If readFromCheckpoint is true,
// it will return all the entries from the last checkpoint (if no checkpoint is
// found, it will return an empty slice.)
func (w *WAL) ReadFromOfsset(offset int, readFromCheckPoint bool) ([]*WAL_Entry, error) {
	//get the list of log segment files in the firectory
	files, err := filepath.Glob(filepath.Join(w.directory, segmentPrefix+"**"))
	if err != nil {
		return nil, err
	}

	var entries []*WAL_Entry
	prevCheckPointLogSequenceNo := uint64(0)

	for _, file := range files {
		//Get the segment index from the file name
		segmentIndex, err := strconv.Atoi(strings.TrimPrefix(file, filepath.Join(w.directory, "segment-")))
		if err != nil {
			return entries, err
		}
		if segmentIndex < offset {
			continue
		}

		file, err := os.OpenFile(file, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}

		entries_from_scratch, checkpoint, err := readAllEntriesFromFile(file, readFromCheckPoint)
		if err != nil {
			return entries, err
		}

		if readFromCheckPoint && checkpoint > prevCheckPointLogSequenceNo {
			entries = entries[:0]
			prevCheckPointLogSequenceNo = checkpoint
		}
		entries = append(entries, entries_from_scratch...)
	}

	return entries, nil
}

// Writes out any data in the WAL's in-memory buffer to the segment file. If
// fsync is enabled, it also calls fsync on the segment file. It also resets
// the synchronization timer.
func (w *WAL) Sync() error {
	if err := w.witeBuffer.Flush(); err != nil {
		return err
	}
	if w.shouldSync {
		if err := w.currentSegment.Sync(); err != nil {
			return err
		}
	}

	//Rest the keepSyncing timer, since we just synced
	w.resetTimer()
	return nil
}

func (w *WAL) resetTimer() {
	w.syncTimer.Reset(syncInterval)
}

func (w *WAL) keepSyncing() {
	for {
		select {
		case <-w.syncTimer.C:
			w.lock.Lock()
			err := w.Sync()
			w.lock.Unlock()

			if err != nil {
				log.Printf("Error while sync performing sync: %v", err)
			}
		case <-w.ctx.Done():
			return
		}
	}
}

// Repairs a corrupted WAL by scanning the WAL from the start and reading all
// entries until a corrupted entry is encountered, at which point the file is
// truncated. The function returns the entries that were read before the
// corruption and overwrites the existing WAL file with the repaired entries.
// It checks the CRC of each entry to verify if it is corrupted, and if the CRC
// is invalid, the file is truncated at that point.
func (w *WAL) Repair() ([]*WAL_Entry, error) {
	files, err := filepath.Glob(filepath.Join(w.directory, segmentPrefix+"*"))
	if err != nil {
		return nil, err
	}

	var lastSegmentID int
	if len(files) > 0 {
		// Find the last segment ID
		lastSegmentID, err = findLastSegmentIndexinFiles(files)
		if err != nil {
			return nil, err
		}
	} else {
		log.Fatalf("No log segments found, nothing to repair.")
	}
	// Open the last log segment file
	filePath := filepath.Join(w.directory, fmt.Sprintf("%s%d", segmentPrefix, lastSegmentID))
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Seek to the beginning of the file
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	var entries []*WAL_Entry

	for {
		// Read the size of the next entry.
		var size int32
		if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				// End of file reached, no corruption found.
				return entries, err
			}
			log.Printf("Error while reading entry size: %v", err)
			// Truncate the file at this point.
			if err := w.replaceWithFixedFile(entries); err != nil {
				return entries, err
			}
			return nil, nil
		}

		// Read the entry data.
		data := make([]byte, size)
		if _, err := io.ReadFull(file, data); err != nil {
			// Truncate the file at this point
			if err := w.replaceWithFixedFile(entries); err != nil {
				return entries, err
			}
			return entries, nil
		}

		// Deserialize the entry.
		var entry WAL_Entry
		// if err := w.Unmarshal(data, &entry); err != nil {
		// 	if err := wal.replaceWithFixedFile(entries); err != nil {
		// 		return entries, err
		// 	}
		// 	return entries, nil
		// }

		if !verifyCRC(&entry) {
			log.Printf("CRC mismatch: data may be corrupted")
			// Truncate the file at this point
			if err := w.replaceWithFixedFile(entries); err != nil {
				return entries, err
			}

			return entries, nil
		}

		// Add the entry to the slice.
		entries = append(entries, &entry)
	}
}

// replaceWithFixedFile replaces the existing WAL file with the given entries
// atomically.
func (w *WAL) replaceWithFixedFile(entries []*WAL_Entry) error {
	// Create a temporary file to make the operation look atomic.
	tempFilePath := fmt.Sprintf("%s.tmp", w.currentSegment.Name())
	tempFile, err := os.OpenFile(tempFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Write the entries to the temporary file
	for _, entry := range entries {
		marshaledEntry := MustMarshal(entry)

		size := int32(len(marshaledEntry))
		if err := binary.Write(tempFile, binary.LittleEndian, size); err != nil {
			return err
		}
		_, err := tempFile.Write(marshaledEntry)

		if err != nil {
			return err
		}
	}

	// Close the temporary file
	if err := tempFile.Close(); err != nil {
		return err
	}

	// Rename the temporary file to the original file name
	if err := os.Rename(tempFilePath, w.currentSegment.Name()); err != nil {
		return err
	}

	return nil
}

// Returns the last sequence number in the current log.
func (w *WAL) getLastSequenceNo() (uint64, error) {
	entry, err := w.getLastEntryInLog()
	if err != nil {
		return 0, err
	}

	if entry != nil {
		return entry.GetLogSequenceNumber(), nil
	}
	return 0, nil
}

// getLastEntryInLog iterates through all the entries of the log and returns the
// last entry.
func (w *WAL) getLastEntryInLog() (*WAL_Entry, error) {
	file, err := os.OpenFile(w.currentSegment.Name(), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var previousSize int32
	var offset int64
	var entry *WAL_Entry

	for {
		var size int32
		if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				// End of file reached, read the last entry at the saved offset.
				if offset == 0 {
					return entry, nil
				}

				if _, err := file.Seek(offset, io.SeekStart); err != nil {
					return nil, err
				}

				// Read the entry data.
				data := make([]byte, previousSize)
				if _, err := io.ReadFull(file, data); err != nil {
					return nil, err
				}

				entry, err = unmarshalAndVerifyEntry(data)
				if err != nil {
					return nil, err
				}

				return entry, nil
			}
			return nil, err
		}

		// Get current offset
		offset, err = file.Seek(0, io.SeekCurrent)
		previousSize = size

		if err != nil {
			return nil, err
		}

		// Skip to the next entry.
		if _, err := file.Seek(int64(size), io.SeekCurrent); err != nil {
			return nil, err
		}
	}
}