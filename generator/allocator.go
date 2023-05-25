package generator

type BitsAllocator struct {
	// Total 64 bits
	TotalBits int32

	// Bits for [sign-> second-> workID-> sequence]

	// Bits for sign
	SignBits int32
	// Bits for second
	TimestampBits int32
	// Bits for worker
	WorkerIdBits int32
	// Bits for sequence
	SequenceBits int32

	// Max value for workID & sequence

	MaxDeltaMilliSeconds uint64
	MaxWorkerId          uint64
	MaxSequence          uint64

	// Shift for timestamp & workerId

	TimestampShift int32
	WorkerIdShift  int32
}

func BitsAllocatorConstructor(timestampBits int32, workerIdBits int32, sequenceBits int32) *BitsAllocator {
	var totalBits = 1 + timestampBits + workerIdBits + sequenceBits
	if totalBits != 1<<6 {
		panic("the sum of bits is not 64")
	}
	bitsAllocator := BitsAllocator{}
	bitsAllocator.TotalBits = totalBits
	bitsAllocator.SignBits = 1
	bitsAllocator.TimestampBits, bitsAllocator.WorkerIdBits, bitsAllocator.SequenceBits = timestampBits, workerIdBits, sequenceBits
	bitsAllocator.MaxDeltaMilliSeconds = maxUint64(timestampBits)
	bitsAllocator.MaxWorkerId = maxUint64(workerIdBits)
	bitsAllocator.MaxSequence = maxUint64(sequenceBits)
	bitsAllocator.TimestampShift = workerIdBits + sequenceBits
	bitsAllocator.WorkerIdShift = sequenceBits
	return &bitsAllocator
}

func maxUint64(bits int32) uint64 {
	temp := -1 ^ (-1 << bits)
	return uint64(temp)
}

func (c *BitsAllocator) Allocate(deltaSeconds uint64, workerId uint64, sequence uint64) uint64 {
	return (deltaSeconds << c.TimestampShift) | (workerId << c.WorkerIdShift) | sequence
}
