package main

import (
	"sync"
	"testing"
	"time"
)

type rowSink struct {
	mu   sync.Mutex
	rows []DetectRow
}

func (s *rowSink) add(rows []DetectRow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows = append(s.rows, rows...)
}

func (s *rowSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.rows)
}

func TestDetectRowBatcherFlushesWhenFull(t *testing.T) {
	sink := &rowSink{}
	batcher := newDetectRowBatcher(sink.add)
	defer batcher.Close()

	for id := 1; id <= detectRowBatchSize; id++ {
		batcher.Add(DetectRow{ID: id, Type: TypePollution})
	}

	if got := sink.count(); got != detectRowBatchSize {
		t.Fatalf("攒满一批应立即发送全部行，实际发送 %d 行", got)
	}
}

func TestDetectRowBatcherFlushesOnInterval(t *testing.T) {
	sink := &rowSink{}
	batcher := newDetectRowBatcher(sink.add)
	defer batcher.Close()

	batcher.Add(DetectRow{ID: 1, Type: TypePollution})

	deadline := time.Now().Add(3 * detectRowFlushInterval)
	for time.Now().Before(deadline) {
		if sink.count() == 1 {
			return
		}
		time.Sleep(detectRowFlushInterval / 4)
	}
	t.Fatal("定时器没有把不足一批的行发出去")
}

func TestDetectRowBatcherFlushesRemainderOnClose(t *testing.T) {
	sink := &rowSink{}
	batcher := newDetectRowBatcher(sink.add)

	batcher.Add(DetectRow{ID: 1, Type: TypePollution})
	batcher.Add(DetectRow{ID: 2, Type: TypePollution})
	batcher.Close()

	if got := sink.count(); got != 2 {
		t.Fatalf("Close 应发送剩余行，实际发送 %d 行", got)
	}
}

func TestDetectRowBatcherCloseIsIdempotent(t *testing.T) {
	sink := &rowSink{}
	batcher := newDetectRowBatcher(sink.add)

	batcher.Add(DetectRow{ID: 1, Type: TypePollution})
	batcher.Close()
	batcher.Close()

	if got := sink.count(); got != 1 {
		t.Fatalf("重复 Close 不应重复发送，实际发送 %d 行", got)
	}
}
