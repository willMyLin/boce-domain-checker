package main

import (
	"sync"
	"time"
)

const (
	// 一批攒够这么多行就立即推给界面，不用等定时器。
	detectRowBatchSize = 100
	// 不足一批时最多等这么久，保证进度条不会停着不动。
	detectRowFlushInterval = 100 * time.Millisecond
)

// detectRowBatcher 把逐行产出的检测结果攒成小批再推给界面。
// 大批量检测时逐行 EventsEmit 会产生数万次跨端调用，界面会被高频消息拖垮；
// 攒批后跨端次数降到百级，前端也能按批一次性合并渲染。
type detectRowBatcher struct {
	emit func([]DetectRow)

	mu   sync.Mutex
	buf  []DetectRow
	done chan struct{}
	once sync.Once
	wg   sync.WaitGroup
}

func newDetectRowBatcher(emit func([]DetectRow)) *detectRowBatcher {
	batcher := &detectRowBatcher{
		emit: emit,
		done: make(chan struct{}),
	}
	batcher.wg.Add(1)
	go batcher.run()
	return batcher
}

func (b *detectRowBatcher) run() {
	defer b.wg.Done()
	ticker := time.NewTicker(detectRowFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flush()
		case <-b.done:
			return
		}
	}
}

// Add 记录一行。攒满一批立即发送，不等定时器。
func (b *detectRowBatcher) Add(row DetectRow) {
	b.mu.Lock()
	b.buf = append(b.buf, row)
	ready := len(b.buf) >= detectRowBatchSize
	b.mu.Unlock()

	if ready {
		b.flush()
	}
}

// Close 停止定时器并发送剩余行。可重复调用。
func (b *detectRowBatcher) Close() {
	b.once.Do(func() {
		close(b.done)
		b.wg.Wait()
		b.flush()
	})
}

func (b *detectRowBatcher) flush() {
	b.mu.Lock()
	if len(b.buf) == 0 {
		b.mu.Unlock()
		return
	}
	rows := b.buf
	b.buf = nil
	b.mu.Unlock()

	b.emit(rows)
}
