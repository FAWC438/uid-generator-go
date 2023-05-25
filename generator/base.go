package generator

import (
	"strconv"
	"sync"
	"time"
	"uid-generator-go/util"
)

type BaseGenerator struct {
	timeBits   int32
	workerBits int32
	seqBits    int32

	epochStr          string
	epochMilliSeconds uint64

	bitsAllocator *BitsAllocator

	workID          uint64
	sequence        uint64
	lastMilliSecond uint64

	IsConstructed bool
	lock          sync.Mutex
}

func BaseGeneratorConstructor() *BaseGenerator {
	generator := BaseGenerator{}
	generator.timeBits, generator.workerBits, generator.seqBits = 41, 10, 12
	generator.epochStr = "2016-05-20"
	//generator.epochMilliSeconds = 1451577600 // ms 2016-01-01 00:00:00
	t, e := time.Parse("2006-01-02", generator.epochStr)
	if e != nil {
		panic("time.Parse error")
	}
	generator.epochMilliSeconds = uint64(t.UnixMilli())
	generator.sequence, generator.lastMilliSecond = 0, 0
	generator.bitsAllocator = BitsAllocatorConstructor(generator.timeBits, generator.workerBits, generator.seqBits)
	generator.workID = 1 // TODO: get it from config or Database
	generator.IsConstructed = true
	generator.lock = sync.Mutex{}
	return &generator
}

func (c *BaseGenerator) GetUID() (uint64, error) {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	uid, e := c.nextID()
	if e != nil {
		return 0, e
	}
	return uid, nil
}

func (c *BaseGenerator) ParseUID(uid uint64) string {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	totalBits, signBits, timestampBits, workerIdBits, sequenceBits :=
		c.bitsAllocator.TotalBits, c.bitsAllocator.SignBits,
		c.bitsAllocator.TimestampBits, c.bitsAllocator.WorkerIdBits, c.bitsAllocator.SequenceBits

	sequence := (uid << (totalBits - sequenceBits)) >> (totalBits - sequenceBits)
	workerId := (uid << (timestampBits + signBits)) >> (totalBits - workerIdBits)
	deltaMilliSeconds := uint64(uid >> (workerIdBits + sequenceBits))
	timestamp := c.epochMilliSeconds + deltaMilliSeconds
	return "UID:" + strconv.FormatUint(uid, 10) + ", timestamp:" + time.UnixMilli(int64(timestamp)).Format("2006-01-02") +
		", workerId:" + strconv.FormatUint(workerId, 10) + ", sequence:" + strconv.FormatUint(sequence, 10)
}

func (c *BaseGenerator) getCurrentMillisecond() (uint64, error) {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	currentMillisecond := uint64(time.Now().UnixMilli())
	if currentMillisecond-c.epochMilliSeconds > c.bitsAllocator.MaxDeltaMilliSeconds {
		return 0, &util.UidGeneratorError{msg: "Timestamp bits is exhausted. Refusing UID generate. Now: " +
			strconv.FormatUint(currentMillisecond, 10) + ", epochSecond: " + strconv.FormatUint(c.epochMilliSeconds, 10)}
	}
	return currentMillisecond, nil
}

func (c *BaseGenerator) getNextMilliSecond(lastMilliSecond uint64) (uint64, error) {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	millisecond, e := c.getCurrentMillisecond()
	if e != nil {
		return 0, e
	}
	for millisecond <= lastMilliSecond {
		millisecond, e = c.getCurrentMillisecond()
		if e != nil {
			return 0, e
		}
	}
	return millisecond, nil
}

func (c *BaseGenerator) nextID() (uint64, error) {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	currentMilliSecond, e := c.getCurrentMillisecond()
	if e != nil {
		return 0, e
	}
	if currentMilliSecond < c.lastMilliSecond {
		return 0, &util.UidGeneratorError{msg: "Clock moved backwards. Refusing for " + strconv.FormatUint(c.lastMilliSecond-currentMilliSecond, 10) + " seconds"}
	}

	if currentMilliSecond == c.lastMilliSecond {
		c.sequence = (c.sequence + 1) & c.bitsAllocator.MaxSequence
		if c.sequence == 0 {
			currentMilliSecond, e = c.getNextMilliSecond(c.lastMilliSecond)
			if e != nil {
				return 0, e
			}
		}
	} else {
		c.sequence = 0
	}
	c.lastMilliSecond = currentMilliSecond
	return c.bitsAllocator.Allocate(currentMilliSecond-c.epochMilliSeconds, c.workID, c.sequence), nil
}

func (c *BaseGenerator) SetWorkID(newWorkID uint64) error {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	if newWorkID < 0 || newWorkID >= 1<<c.bitsAllocator.WorkerIdBits {
		return &util.UidGeneratorError{msg: "Set WorkID error"}
	}
	c.workID = newWorkID
	return nil
}

func (c *BaseGenerator) SetEpochStr(newEpochStr string) error {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	t, e := time.Parse("2006-01-02", newEpochStr)
	if e != nil {
		return &util.UidGeneratorError{msg: "Set epochStr error"}
	}
	c.epochStr = newEpochStr
	c.epochMilliSeconds = uint64(t.UnixMilli())
	return nil
}

func (c *BaseGenerator) SetTimeBits(newTimeBits int32) error {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	if newTimeBits < 0 || newTimeBits >= 64 {
		return &util.UidGeneratorError{msg: "Set timeBits error"}
	}
	c.timeBits = newTimeBits
	c.bitsAllocator = BitsAllocatorConstructor(c.timeBits, c.workerBits, c.seqBits)
	return nil
}

func (c *BaseGenerator) SetWorkerBits(newWorkerBits int32) error {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	if newWorkerBits < 0 || newWorkerBits >= 64 {
		return &util.UidGeneratorError{msg: "Set workerBits error"}
	}
	c.workerBits = newWorkerBits
	c.bitsAllocator = BitsAllocatorConstructor(c.timeBits, c.workerBits, c.seqBits)
	return nil
}

func (c *BaseGenerator) SetSeqBits(newSeqBits int32) error {
	if !c.IsConstructed {
		panic("BaseGenerator is not constructed")
	}
	if newSeqBits < 0 || newSeqBits >= 64 {
		return &util.UidGeneratorError{msg: "Set seqBits error"}
	}
	c.seqBits = newSeqBits
	c.bitsAllocator = BitsAllocatorConstructor(c.timeBits, c.workerBits, c.seqBits)
	return nil
}
