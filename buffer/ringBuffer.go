package buffer

import (
	"strconv"
	"sync"
	"uid-generator-go/util"
)

var startPoint int64 = -1
var enablePutFlag int64 = 0
var enableTakeFlag int64 = 1
var DefaultPaddingPercent int32 = 50

type RingBuffer struct {
	bufferSize int32
	indexMask  int64
	slots      []int64

	flags            []util.PaddedAtomicLong
	tail             *util.PaddedAtomicLong
	cursor           *util.PaddedAtomicLong
	paddingThreshold int32

	lock sync.Mutex
}

func NewRingBuffer(bufferSize int32, paddingFactor int32) *RingBuffer {
	if bufferSize < 1 {
		panic("bufferSize must be positive")
	}
	if bufferSize&(bufferSize-1) != 0 {
		panic("bufferSize must be a power of 2")
	}
	if paddingFactor < 0 || paddingFactor > 100 {
		panic("paddingFactor must be between 0 and 100")
	}
	ringBuffer := RingBuffer{}
	ringBuffer.tail = util.NewPaddedAtomicLong(startPoint)
	ringBuffer.cursor = util.NewPaddedAtomicLong(startPoint)
	ringBuffer.bufferSize = bufferSize
	ringBuffer.indexMask = int64(bufferSize - 1)
	ringBuffer.slots = make([]int64, bufferSize)
	ringBuffer.flags = make([]util.PaddedAtomicLong, bufferSize)
	for i := int32(0); i < bufferSize; i++ {
		ringBuffer.flags[i] = *util.NewPaddedAtomicLong(enablePutFlag)
	}
	ringBuffer.paddingThreshold = bufferSize * paddingFactor / 100
	ringBuffer.lock = sync.Mutex{}

	return &ringBuffer
}

func (c *RingBuffer) Put(uid int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	currentTail := c.tail.Value.Load()
	currentCursor := c.cursor.Value.Load()
	distance := currentTail

	if currentCursor != startPoint {
		distance = currentTail - currentCursor
	}
	if distance == int64(c.bufferSize-1) {
		return &util.BufferFullError{Msg: "Rejected putting buffer for uid:" + strconv.FormatInt(uid, 10)}
	}

	nextTailIndex := (currentTail + 1) & c.indexMask
	if c.flags[nextTailIndex].Value.Load() != enablePutFlag {
		return &util.BufferFullError{Msg: "Rejected putting buffer for uid:" + strconv.FormatInt(uid, 10)}
	}

	c.slots[nextTailIndex] = uid
	c.flags[nextTailIndex].Value.Store(enableTakeFlag)
	c.tail.Value.Add(1)

	return nil
}

func (c *RingBuffer) Take() (int64, error) {
	currentCursor := c.cursor.Value.Load()
	nextCursor := currentCursor
	oldTail := c.tail.Value.Load()
	// TODO: Is there a better way to do this?
	if oldTail != currentCursor {
		nextCursor = c.cursor.Value.Add(1)
	}
	if nextCursor < currentCursor {
		return 0, &util.UidGeneratorError{Msg: "Cursor can't move back"}
	}

	currentTail := c.tail.Value.Load()
	if currentTail-nextCursor < int64(c.paddingThreshold) {
		// TODO: Use goroutine pool to do this
	}

	if currentTail == nextCursor {
		return 0, &util.UidGeneratorError{Msg: "Rejected take buffer"}
	}

	nextCursorIndex := nextCursor & c.indexMask
	if c.flags[nextCursorIndex].Value.Load() != enableTakeFlag {
		return 0, &util.UidGeneratorError{Msg: "Cursor not in can take status"}
	}

	uid := c.slots[nextCursorIndex]
	c.flags[nextCursorIndex].Value.Store(enablePutFlag)

	return uid, nil
}
