package writeaheadlog

import (
	"bufio"
	"context"
	"os"
	"sync"
	"time"
)

type WAL struct {
	path                string
	segmentSize         int
	segmentEntries      int
	IndexSpace          int
	logSuffix           string
	indexSuffix         string
	base                int
	maxSegments         int
	noSplitSegement     bool
	nameLength          int
	close               bool
	segments            []*segments
	firstIndex          uint64
	lastIndex           uint64
	lastSegement        []*segments
	encoderBuffer       []byte
	witeBuffer          bufio.Writer
	directory           string
	lock                sync.Mutex
	lastSequenceNo      uint64
	syncTimer           *time.Timer
	shouldSync         bool
	maxFileSize         int64
	ctx                 context.Context
	cancel              context.CancelFunc
	currentSegment *os.File
	currentSegmentIndex int
}

type segments struct {
	maxSegments         int
	logPath        string
	indexPath      string
	indexSpace     int
	offset         uint64
	len            uint64
	currentSegment *os.File
	currentSegmentIndex int
	indexFile      *os.File
	indexMmap      []byte
	logFile        *os.File
	indexBuffer    []byte
}


type WAL_Entry struct{
    logSequenceNumber int
    datan []byte
    CRC uint64
	isCheckPoint bool
}