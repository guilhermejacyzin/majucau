package csvstream

import (
	"bufio"
	"container/heap"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"majucau.local/financial-intelligence/internal/securetemp"
)

const (
	diskInitialCapacity = uint64(1024)
	diskSlotSize        = int64(1 + sha256.Size + 8)
	diskMergeFanIn      = 64
	maxDiskNameBytes    = 1 << 20
	nameNonceSize       = 12
	maxDiskInt64        = int64(^uint64(0) >> 1)
	maxDiskUint64       = ^uint64(0)
)

var (
	queueNameAAD    = []byte("majucau/csv-preview/ordered-name/v1")
	locationNameAAD = []byte("majucau/csv-preview/source-location/v1")
)

// DiskSourceIDSet is an exact, temporary deduplication index for previews.
// It stores keyed HMACs and encrypted source locations under the operating
// system's private temporary directory; the original source identifiers are
// never written to disk. Its hash table and filename sort runs remain on disk,
// keeping heap use bounded as row and file counts grow.
type DiskSourceIDSet struct {
	dir          string
	indexPath    string
	index        *os.File
	location     *os.File
	capacity     uint64
	count        uint64
	hmacKey      [32]byte
	gcm          cipher.AEAD
	noncePrefix  [4]byte
	nonceCounter uint64
	closed       bool
}

func NewDiskSourceIDSet(ctx context.Context) (*DiskSourceIDSet, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dir, err := securetemp.NewDir(os.TempDir(), "majucau-csv-preview-")
	if err != nil {
		return nil, fmt.Errorf("create private CSV preview workspace: %w", err)
	}
	set := &DiskSourceIDSet{dir: dir, capacity: diskInitialCapacity}
	cleanup := func(cause error) (*DiskSourceIDSet, error) {
		return nil, errors.Join(cause, set.Close())
	}
	var encryptionKey [32]byte
	defer func() {
		for index := range encryptionKey {
			encryptionKey[index] = 0
		}
	}()
	if _, err := rand.Read(set.hmacKey[:]); err != nil {
		return cleanup(fmt.Errorf("generate CSV preview HMAC key: %w", err))
	}
	if _, err := rand.Read(encryptionKey[:]); err != nil {
		return cleanup(fmt.Errorf("generate CSV preview encryption key: %w", err))
	}
	if _, err := rand.Read(set.noncePrefix[:]); err != nil {
		return cleanup(fmt.Errorf("generate CSV preview nonce prefix: %w", err))
	}
	block, err := aes.NewCipher(encryptionKey[:])
	for index := range encryptionKey {
		encryptionKey[index] = 0
	}
	if err != nil {
		return cleanup(fmt.Errorf("initialize CSV preview encryption: %w", err))
	}
	set.gcm, err = cipher.NewGCM(block)
	if err != nil {
		return cleanup(fmt.Errorf("initialize CSV preview encryption: %w", err))
	}
	set.index, err = os.CreateTemp(dir, "seen-*.idx")
	if err != nil {
		return cleanup(fmt.Errorf("create temporary CSV source index: %w", err))
	}
	set.indexPath = set.index.Name()
	if err := set.index.Truncate(int64(set.capacity) * diskSlotSize); err != nil {
		return cleanup(fmt.Errorf("initialize temporary CSV source index: %w", err))
	}
	set.location, err = os.CreateTemp(dir, "locations-*.enc")
	if err != nil {
		return cleanup(fmt.Errorf("create temporary CSV location store: %w", err))
	}
	return set, nil
}

func (s *DiskSourceIDSet) CheckAndAdd(ctx context.Context, sourceID, file string, line int) (string, int64, bool, error) {
	if s == nil || s.closed || s.index == nil || s.location == nil || sourceID == "" {
		return "", 0, false, errors.New("temporary CSV source index is not initialized")
	}
	if err := ctx.Err(); err != nil {
		return "", 0, false, err
	}
	if len(file) > maxDiskNameBytes {
		return "", 0, false, errors.New("CSV filename exceeds temporary index limit")
	}
	if s.count == maxDiskUint64 {
		return "", 0, false, errors.New("temporary CSV source index capacity exceeded")
	}
	threshold := s.capacity/10*7 + s.capacity%10*7/10
	if s.count+1 > threshold {
		if err := s.grow(); err != nil {
			return "", 0, false, err
		}
	}
	mac := hmac.New(sha256.New, s.hmacKey[:])
	sourceBytes := []byte(sourceID)
	_, _ = mac.Write(sourceBytes)
	for index := range sourceBytes {
		sourceBytes[index] = 0
	}
	fingerprint := mac.Sum(nil)
	position, err := initialDiskPosition(fingerprint, s.capacity)
	if err != nil {
		return "", 0, false, err
	}
	for probes := uint64(0); probes < s.capacity; probes++ {
		if err := ctx.Err(); err != nil {
			return "", 0, false, err
		}
		slot, err := s.readSlot(position)
		if err != nil {
			return "", 0, false, err
		}
		if slot[0] == 0 {
			locationOffset, err := s.appendLocation(file, line)
			if err != nil {
				return "", 0, false, err
			}
			slot[0] = 1
			copy(slot[1:1+sha256.Size], fingerprint)
			binary.BigEndian.PutUint64(slot[1+sha256.Size:], locationOffset)
			if err := s.writeSlot(position, slot); err != nil {
				return "", 0, false, err
			}
			s.count++
			return "", 0, false, nil
		}
		if hmac.Equal(slot[1:1+sha256.Size], fingerprint) {
			locationOffset := binary.BigEndian.Uint64(slot[1+sha256.Size:])
			firstFile, firstLine, err := s.readLocation(locationOffset)
			if err != nil {
				return "", 0, false, err
			}
			return firstFile, firstLine, true, nil
		}
		position = (position + 1) & (s.capacity - 1)
	}
	return "", 0, false, errors.New("temporary CSV source index is full")
}

func (s *DiskSourceIDSet) QueueCSVFiles(ctx context.Context, folder string) (CSVFileQueue, error) {
	if s == nil || s.closed || s.dir == "" || s.gcm == nil {
		return nil, errors.New("temporary CSV source index is not initialized")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	directory, err := os.Open(folder)
	if err != nil {
		return nil, fmt.Errorf("open CSV import folder: %w", err)
	}
	defer directory.Close()
	builder := diskRunBuilder{set: s}
	ignoredCount := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entries, readErr := directory.ReadDir(directoryReadBatchSize)
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("inspect CSV import folder entry: %w", err)
			}
			if !info.Mode().IsRegular() || !strings.EqualFold(filepath.Ext(entry.Name()), ".csv") {
				ignoredCount++
				continue
			}
			names = append(names, entry.Name())
		}
		if len(names) > 0 {
			sort.Slice(names, func(i, j int) bool { return diskNameLess(names[i], names[j]) })
			run, err := s.writeNameRun(names)
			if err != nil {
				return nil, err
			}
			if err := builder.add(ctx, run); err != nil {
				return nil, err
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read CSV import folder entries: %w", readErr)
		}
	}
	if err := directory.Close(); err != nil {
		return nil, fmt.Errorf("close CSV import folder: %w", err)
	}
	finalPath, err := builder.finish(ctx)
	if err != nil {
		return nil, err
	}
	queue := &diskFileQueue{ctx: ctx, set: s, ignoredCount: ignoredCount}
	if finalPath != "" {
		queue.file, err = os.Open(finalPath)
		if err != nil {
			return nil, fmt.Errorf("open ordered temporary CSV filename queue: %w", err)
		}
		queue.reader = bufio.NewReaderSize(queue.file, 32*1024)
	}
	return queue, nil
}

func (s *DiskSourceIDSet) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	var closeErrors []error
	if s.index != nil {
		if err := s.index.Close(); err != nil {
			closeErrors = append(closeErrors, err)
		}
		s.index = nil
	}
	if s.location != nil {
		if err := s.location.Close(); err != nil {
			closeErrors = append(closeErrors, err)
		}
		s.location = nil
	}
	for index := range s.hmacKey {
		s.hmacKey[index] = 0
	}
	for index := range s.noncePrefix {
		s.noncePrefix[index] = 0
	}
	s.gcm = nil
	if s.dir != "" {
		if err := os.RemoveAll(s.dir); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("remove temporary CSV preview workspace: %w", err))
		}
		s.dir = ""
	}
	return errors.Join(closeErrors...)
}

func (s *DiskSourceIDSet) readSlot(position uint64) ([]byte, error) {
	data := make([]byte, diskSlotSize)
	if _, err := s.index.ReadAt(data, int64(position)*diskSlotSize); err != nil {
		return nil, fmt.Errorf("read temporary CSV source index: %w", err)
	}
	return data, nil
}

func (s *DiskSourceIDSet) writeSlot(position uint64, data []byte) error {
	if len(data) != int(diskSlotSize) {
		return errors.New("invalid temporary CSV source index slot")
	}
	if _, err := s.index.WriteAt(data, int64(position)*diskSlotSize); err != nil {
		return fmt.Errorf("write temporary CSV source index: %w", err)
	}
	return nil
}

func (s *DiskSourceIDSet) appendLocation(file string, line int) (uint64, error) {
	if line < 0 {
		return 0, errors.New("invalid CSV source line")
	}
	plaintext := make([]byte, 8+len(file))
	binary.BigEndian.PutUint64(plaintext[:8], uint64(line))
	copy(plaintext[8:], file)
	sealed, err := s.sealName(plaintext, locationNameAAD)
	for index := range plaintext {
		plaintext[index] = 0
	}
	if err != nil {
		return 0, err
	}
	position, err := s.location.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("locate temporary CSV source location: %w", err)
	}
	if err := writeSealedRecord(s.location, sealed); err != nil {
		return 0, fmt.Errorf("write temporary CSV source location: %w", err)
	}
	return uint64(position), nil
}

func (s *DiskSourceIDSet) readLocation(position uint64) (string, int64, error) {
	if position > uint64(maxDiskInt64) {
		return "", 0, errors.New("invalid temporary CSV source location offset")
	}
	plaintext, err := readSealedRecordAt(s.location, int64(position), s.gcm, locationNameAAD, maxDiskNameBytes+8)
	if err != nil {
		return "", 0, fmt.Errorf("read temporary CSV source location: %w", err)
	}
	if len(plaintext) < 8 {
		for index := range plaintext {
			plaintext[index] = 0
		}
		return "", 0, errors.New("invalid temporary CSV source location")
	}
	line := binary.BigEndian.Uint64(plaintext[:8])
	if line > uint64(maxDiskInt64) {
		for index := range plaintext {
			plaintext[index] = 0
		}
		return "", 0, errors.New("invalid temporary CSV source line")
	}
	file := string(plaintext[8:])
	for index := range plaintext {
		plaintext[index] = 0
	}
	return file, int64(line), nil
}

func (s *DiskSourceIDSet) sealName(plaintext, aad []byte) ([]byte, error) {
	if s == nil || s.gcm == nil || len(plaintext) > maxDiskNameBytes+8 {
		return nil, errors.New("temporary CSV encrypted value exceeds limit")
	}
	if s.nonceCounter == maxDiskUint64 {
		return nil, errors.New("temporary CSV encryption nonce capacity exceeded")
	}
	s.nonceCounter++
	var nonce [nameNonceSize]byte
	copy(nonce[:4], s.noncePrefix[:])
	binary.BigEndian.PutUint64(nonce[4:], s.nonceCounter)
	sealed := s.gcm.Seal(nonce[:], nonce[:], plaintext, aad)
	return sealed, nil
}

func (s *DiskSourceIDSet) writeNameRun(names []string) (string, error) {
	file, err := os.CreateTemp(s.dir, "names-*.run")
	if err != nil {
		return "", fmt.Errorf("create temporary ordered filename run: %w", err)
	}
	path := file.Name()
	writer := bufio.NewWriterSize(file, 32*1024)
	for _, name := range names {
		if err := s.writeQueueName(writer, name); err != nil {
			_ = file.Close()
			return "", fmt.Errorf("write temporary ordered filename run: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("flush temporary ordered filename run: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close temporary ordered filename run: %w", err)
	}
	return path, nil
}

func (s *DiskSourceIDSet) writeQueueName(writer io.Writer, name string) error {
	if len(name) > maxDiskNameBytes {
		return errors.New("CSV filename exceeds temporary queue limit")
	}
	plaintext := []byte(name)
	sealed, err := s.sealName(plaintext, queueNameAAD)
	for index := range plaintext {
		plaintext[index] = 0
	}
	if err != nil {
		return err
	}
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(sealed)))
	if _, err := writer.Write(length[:]); err != nil {
		return err
	}
	_, err = writer.Write(sealed)
	return err
}

func (s *DiskSourceIDSet) readQueueName(reader *bufio.Reader) (string, error) {
	var length [4]byte
	n, err := io.ReadFull(reader, length[:])
	if errors.Is(err, io.EOF) && n == 0 {
		return "", io.EOF
	}
	if err != nil {
		return "", fmt.Errorf("read temporary CSV filename length: %w", err)
	}
	size := binary.BigEndian.Uint32(length[:])
	minimum := uint32(nameNonceSize + s.gcm.Overhead())
	maximum := uint32(maxDiskNameBytes + nameNonceSize + s.gcm.Overhead())
	if size < minimum || size > maximum {
		return "", errors.New("invalid temporary CSV filename record")
	}
	sealed := make([]byte, int(size))
	if _, err := io.ReadFull(reader, sealed); err != nil {
		return "", fmt.Errorf("read temporary CSV filename record: %w", err)
	}
	plaintext, err := s.gcm.Open(nil, sealed[:nameNonceSize], sealed[nameNonceSize:], queueNameAAD)
	if err != nil {
		return "", errors.New("authenticate temporary CSV filename record")
	}
	name := string(plaintext)
	for index := range plaintext {
		plaintext[index] = 0
	}
	return name, nil
}

func (s *DiskSourceIDSet) grow() error {
	if s.capacity > uint64(maxDiskInt64/diskSlotSize)/2 {
		return errors.New("temporary CSV source index capacity exceeded")
	}
	newCapacity := s.capacity * 2
	newIndex, err := os.CreateTemp(s.dir, "seen-resize-*.idx")
	if err != nil {
		return fmt.Errorf("grow temporary CSV source index: %w", err)
	}
	newPath := newIndex.Name()
	if err := newIndex.Truncate(int64(newCapacity) * diskSlotSize); err != nil {
		_ = newIndex.Close()
		_ = os.Remove(newPath)
		return fmt.Errorf("initialize larger temporary CSV source index: %w", err)
	}
	for oldPosition := uint64(0); oldPosition < s.capacity; oldPosition++ {
		slot, err := s.readSlot(oldPosition)
		if err != nil {
			_ = newIndex.Close()
			_ = os.Remove(newPath)
			return err
		}
		if slot[0] == 0 {
			continue
		}
		fingerprint := slot[1 : 1+sha256.Size]
		position, err := initialDiskPosition(fingerprint, newCapacity)
		if err != nil {
			_ = newIndex.Close()
			_ = os.Remove(newPath)
			return err
		}
		for {
			other := make([]byte, diskSlotSize)
			if _, err := newIndex.ReadAt(other, int64(position)*diskSlotSize); err != nil {
				_ = newIndex.Close()
				_ = os.Remove(newPath)
				return fmt.Errorf("read larger temporary CSV source index: %w", err)
			}
			if other[0] == 0 {
				if _, err := newIndex.WriteAt(slot, int64(position)*diskSlotSize); err != nil {
					_ = newIndex.Close()
					_ = os.Remove(newPath)
					return fmt.Errorf("rehash temporary CSV source index: %w", err)
				}
				break
			}
			position = (position + 1) & (newCapacity - 1)
		}
	}
	oldIndex, oldPath := s.index, s.indexPath
	if err := oldIndex.Close(); err != nil {
		_ = newIndex.Close()
		_ = os.Remove(newPath)
		return fmt.Errorf("close old temporary CSV source index: %w", err)
	}
	if err := os.Remove(oldPath); err != nil {
		_ = newIndex.Close()
		return fmt.Errorf("replace old temporary CSV source index: %w", err)
	}
	s.index = newIndex
	s.indexPath = newPath
	s.capacity = newCapacity
	return nil
}

func initialDiskPosition(fingerprint []byte, capacity uint64) (uint64, error) {
	if len(fingerprint) != sha256.Size || capacity == 0 || capacity&(capacity-1) != 0 {
		return 0, errors.New("invalid temporary CSV source index parameters")
	}
	hash := fnv.New64a()
	_, _ = hash.Write(fingerprint)
	return hash.Sum64() & (capacity - 1), nil
}

func writeSealedRecord(writer io.Writer, sealed []byte) error {
	if len(sealed) < nameNonceSize {
		return errors.New("invalid encrypted temporary CSV record")
	}
	var header [16]byte
	copy(header[:nameNonceSize], sealed[:nameNonceSize])
	binary.BigEndian.PutUint32(header[nameNonceSize:], uint32(len(sealed)-nameNonceSize))
	if _, err := writer.Write(header[:]); err != nil {
		return err
	}
	_, err := writer.Write(sealed[nameNonceSize:])
	return err
}

func readSealedRecordAt(file *os.File, offset int64, gcm cipher.AEAD, aad []byte, maxPlaintext int) ([]byte, error) {
	var header [16]byte
	if _, err := file.ReadAt(header[:], offset); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header[nameNonceSize:])
	if uint64(length) > uint64(maxPlaintext+gcm.Overhead()) {
		return nil, errors.New("invalid encrypted temporary CSV record size")
	}
	ciphertext := make([]byte, int(length))
	if _, err := file.ReadAt(ciphertext, offset+int64(len(header))); err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, header[:nameNonceSize], ciphertext, aad)
	if err != nil {
		return nil, errors.New("authenticate temporary CSV source location")
	}
	return plaintext, nil
}

func diskNameLess(left, right string) bool {
	leftKey, rightKey := strings.ToLower(left), strings.ToLower(right)
	if leftKey == rightKey {
		return left < right
	}
	return leftKey < rightKey
}

type diskRunBuilder struct {
	set    *DiskSourceIDSet
	levels [][]string
}

func (b *diskRunBuilder) add(ctx context.Context, path string) error {
	for level := 0; ; level++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		for len(b.levels) <= level {
			b.levels = append(b.levels, nil)
		}
		b.levels[level] = append(b.levels[level], path)
		if len(b.levels[level]) < diskMergeFanIn {
			return nil
		}
		group := b.levels[level]
		merged, err := b.set.mergeNameRuns(ctx, group)
		if err != nil {
			return err
		}
		for _, oldPath := range group {
			if err := os.Remove(oldPath); err != nil {
				return fmt.Errorf("remove merged temporary CSV filename run: %w", err)
			}
		}
		b.levels[level] = nil
		path = merged
	}
}

func (b *diskRunBuilder) finish(ctx context.Context) (string, error) {
	paths := make([]string, 0, len(b.levels)*diskMergeFanIn)
	for _, level := range b.levels {
		paths = append(paths, level...)
	}
	for len(paths) > 1 {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		next := make([]string, 0, (len(paths)+diskMergeFanIn-1)/diskMergeFanIn)
		for start := 0; start < len(paths); start += diskMergeFanIn {
			end := min(start+diskMergeFanIn, len(paths))
			group := paths[start:end]
			if len(group) == 1 {
				next = append(next, group[0])
				continue
			}
			merged, err := b.set.mergeNameRuns(ctx, group)
			if err != nil {
				return "", err
			}
			for _, oldPath := range group {
				if err := os.Remove(oldPath); err != nil {
					return "", fmt.Errorf("remove merged temporary CSV filename run: %w", err)
				}
			}
			next = append(next, merged)
		}
		paths = next
	}
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], nil
}

func (s *DiskSourceIDSet) mergeNameRuns(ctx context.Context, paths []string) (string, error) {
	if len(paths) == 0 || len(paths) > diskMergeFanIn {
		return "", errors.New("invalid temporary CSV filename merge group")
	}
	output, err := os.CreateTemp(s.dir, "names-merged-*.run")
	if err != nil {
		return "", fmt.Errorf("create merged temporary CSV filename run: %w", err)
	}
	outputPath := output.Name()
	removeOutput := true
	defer func() {
		if removeOutput {
			_ = output.Close()
			_ = os.Remove(outputPath)
		}
	}()
	writer := bufio.NewWriterSize(output, 32*1024)
	readers := make([]*diskRunReader, 0, len(paths))
	defer func() {
		for _, reader := range readers {
			_ = reader.file.Close()
		}
	}()
	items := make(diskRunHeap, 0, len(paths))
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		file, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("open temporary CSV filename run: %w", err)
		}
		reader := &diskRunReader{file: file, reader: bufio.NewReaderSize(file, 32*1024)}
		readers = append(readers, reader)
		name, err := s.readQueueName(reader.reader)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			return "", err
		}
		items = append(items, diskRunItem{name: name, sortKey: strings.ToLower(name), reader: reader})
	}
	heap.Init(&items)
	for items.Len() > 0 {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		item := heap.Pop(&items).(diskRunItem)
		if err := s.writeQueueName(writer, item.name); err != nil {
			return "", fmt.Errorf("write merged temporary CSV filename run: %w", err)
		}
		name, err := s.readQueueName(item.reader.reader)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			return "", err
		}
		item.name = name
		item.sortKey = strings.ToLower(name)
		heap.Push(&items, item)
	}
	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("flush merged temporary CSV filename run: %w", err)
	}
	if err := output.Close(); err != nil {
		return "", fmt.Errorf("close merged temporary CSV filename run: %w", err)
	}
	removeOutput = false
	return outputPath, nil
}

type diskRunReader struct {
	file   *os.File
	reader *bufio.Reader
}

type diskRunItem struct {
	name    string
	sortKey string
	reader  *diskRunReader
}

type diskRunHeap []diskRunItem

func (h diskRunHeap) Len() int { return len(h) }
func (h diskRunHeap) Less(i, j int) bool {
	if h[i].sortKey == h[j].sortKey {
		return h[i].name < h[j].name
	}
	return h[i].sortKey < h[j].sortKey
}
func (h diskRunHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *diskRunHeap) Push(value any) {
	*h = append(*h, value.(diskRunItem))
}
func (h *diskRunHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}

type diskFileQueue struct {
	ctx          context.Context
	set          *DiskSourceIDSet
	file         *os.File
	reader       *bufio.Reader
	ignoredCount int
	closed       bool
}

func (q *diskFileQueue) Next() (string, bool, error) {
	if q == nil || q.closed || q.set == nil {
		return "", false, errors.New("temporary CSV filename queue is not initialized")
	}
	if err := q.ctx.Err(); err != nil {
		return "", false, err
	}
	if q.reader == nil {
		return "", false, nil
	}
	name, err := q.set.readQueueName(q.reader)
	if errors.Is(err, io.EOF) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return name, true, nil
}

func (q *diskFileQueue) IgnoredCount() int {
	if q == nil {
		return 0
	}
	return q.ignoredCount
}

func (q *diskFileQueue) Close() {
	if q == nil || q.closed {
		return
	}
	q.closed = true
	if q.file != nil {
		_ = q.file.Close()
		q.file = nil
	}
	q.reader = nil
	q.set = nil
	q.ctx = nil
}
