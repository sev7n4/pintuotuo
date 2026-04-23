package services

import (
	"context"
	"sync"
	"time"
)

type RequestQueue interface {
	Enqueue(ctx context.Context, req *QueuedRequest) error
	Dequeue(ctx context.Context) (*QueuedRequest, error)
	Size() int
	IsEmpty() bool
	GetStats() map[string]interface{}
}

type QueuedRequest struct {
	RequestID      string                 `json:"request_id"`
	MerchantID     int                    `json:"merchant_id"`
	Model          string                 `json:"model"`
	Provider       string                 `json:"provider,omitempty"`
	RequestBody    map[string]interface{} `json:"request_body"`
	UserPrefs      map[string]interface{} `json:"user_preferences,omitempty"`
	CostBudget     *float64               `json:"cost_budget,omitempty"`
	ComplianceReqs []string               `json:"compliance_requirements,omitempty"`
	AllowedKeyIDs  []int                  `json:"allowed_key_ids,omitempty"`
	Priority       int                    `json:"priority"`
	EnqueuedAt     time.Time              `json:"enqueued_at"`
	Timeout        time.Duration          `json:"timeout"`
}

type PriorityQueue struct {
	mu      sync.Mutex
	reqs    []*QueuedRequest
	stats   *QueueStats
	maxSize int
}

type QueueStats struct {
	Enqueued    int64     `json:"enqueued"`
	Dequeued    int64     `json:"dequeued"`
	Expired     int64     `json:"expired"`
	Dropped     int64     `json:"dropped"`
	AvgWaitTime float64   `json:"avg_wait_time_ms"`
	MaxWaitTime float64   `json:"max_wait_time_ms"`
	LastReset   time.Time `json:"last_reset"`
	mu          sync.Mutex
}

type QueueFactory struct {
	queues map[string]*PriorityQueue
	mu     sync.RWMutex
}

var queueFactory *QueueFactory
var queueOnce sync.Once

func GetQueueFactory() *QueueFactory {
	queueOnce.Do(func() {
		queueFactory = &QueueFactory{
			queues: make(map[string]*PriorityQueue),
		}
	})
	return queueFactory
}

func NewPriorityQueue(maxSize int) *PriorityQueue {
	return &PriorityQueue{
		reqs: make([]*QueuedRequest, 0),
		stats: &QueueStats{
			LastReset: time.Now(),
		},
		maxSize: maxSize,
	}
}

func (pq *PriorityQueue) Enqueue(ctx context.Context, req *QueuedRequest) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// 检查队列大小
	if len(pq.reqs) >= pq.maxSize {
		pq.stats.mu.Lock()
		pq.stats.Dropped++
		pq.stats.mu.Unlock()
		return nil
	}

	// 设置入队时间
	req.EnqueuedAt = time.Now()

	// 按优先级插入
	insertPos := 0
	for insertPos < len(pq.reqs) && pq.reqs[insertPos].Priority > req.Priority {
		insertPos++
	}

	// 插入到正确位置
	pq.reqs = append(pq.reqs, nil)
	copy(pq.reqs[insertPos+1:], pq.reqs[insertPos:])
	pq.reqs[insertPos] = req

	// 更新统计信息
	pq.stats.mu.Lock()
	pq.stats.Enqueued++
	pq.stats.mu.Unlock()

	return nil
}

func (pq *PriorityQueue) Dequeue(ctx context.Context) (*QueuedRequest, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// 清理过期请求
	pq.cleanupExpired()

	if len(pq.reqs) == 0 {
		return nil, nil
	}

	// 取出优先级最高的请求
	req := pq.reqs[0]
	pq.reqs = pq.reqs[1:]

	// 计算等待时间
	waitTime := time.Since(req.EnqueuedAt).Milliseconds()

	// 更新统计信息
	pq.stats.mu.Lock()
	pq.stats.Dequeued++
	pq.stats.AvgWaitTime = (pq.stats.AvgWaitTime*float64(pq.stats.Dequeued-1) + float64(waitTime)) / float64(pq.stats.Dequeued)
	if float64(waitTime) > pq.stats.MaxWaitTime {
		pq.stats.MaxWaitTime = float64(waitTime)
	}
	pq.stats.mu.Unlock()

	return req, nil
}

func (pq *PriorityQueue) cleanupExpired() {
	now := time.Now()
	expiredCount := 0

	for i, req := range pq.reqs {
		if req.Timeout > 0 && now.Sub(req.EnqueuedAt) > req.Timeout {
			expiredCount++
		} else {
			// 所有后续请求都未过期
			if expiredCount > 0 {
				// 移除过期请求
				pq.reqs = append(pq.reqs[:i-expiredCount], pq.reqs[i:]...)
			}
			break
		}
	}

	// 处理所有请求都过期的情况
	if expiredCount > 0 && expiredCount == len(pq.reqs) {
		pq.reqs = []*QueuedRequest{}
	}

	// 更新统计信息
	if expiredCount > 0 {
		pq.stats.mu.Lock()
		pq.stats.Expired += int64(expiredCount)
		pq.stats.mu.Unlock()
	}
}

func (pq *PriorityQueue) Size() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// 清理过期请求
	pq.cleanupExpired()

	return len(pq.reqs)
}

func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Size() == 0
}

func (pq *PriorityQueue) GetStats() map[string]interface{} {
	pq.mu.Lock()
	// 清理过期请求
	pq.cleanupExpired()
	currentSize := len(pq.reqs)
	pq.mu.Unlock()

	pq.stats.mu.Lock()
	defer pq.stats.mu.Unlock()

	return map[string]interface{}{
		"current_size":     currentSize,
		"max_size":         pq.maxSize,
		"enqueued":         pq.stats.Enqueued,
		"dequeued":         pq.stats.Dequeued,
		"expired":          pq.stats.Expired,
		"dropped":          pq.stats.Dropped,
		"avg_wait_time_ms": pq.stats.AvgWaitTime,
		"max_wait_time_ms": pq.stats.MaxWaitTime,
		"last_reset":       pq.stats.LastReset,
	}
}

func (pq *PriorityQueue) ResetStats() {
	pq.stats.mu.Lock()
	defer pq.stats.mu.Unlock()

	pq.stats.Enqueued = 0
	pq.stats.Dequeued = 0
	pq.stats.Expired = 0
	pq.stats.Dropped = 0
	pq.stats.AvgWaitTime = 0
	pq.stats.MaxWaitTime = 0
	pq.stats.LastReset = time.Now()
}

func (pq *PriorityQueue) SetMaxSize(maxSize int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.maxSize = maxSize

	// 如果队列超过新的最大大小，移除优先级最低的请求
	if len(pq.reqs) > maxSize {
		dropCount := len(pq.reqs) - maxSize
		pq.reqs = pq.reqs[:maxSize]

		// 更新统计信息
		pq.stats.mu.Lock()
		pq.stats.Dropped += int64(dropCount)
		pq.stats.mu.Unlock()
	}
}

func (f *QueueFactory) GetQueue(key string, maxSize int) *PriorityQueue {
	f.mu.Lock()
	defer f.mu.Unlock()

	queue, exists := f.queues[key]
	if !exists {
		queue = NewPriorityQueue(maxSize)
		f.queues[key] = queue
	}

	return queue
}

func (f *QueueFactory) Enqueue(key string, maxSize int, req *QueuedRequest) error {
	queue := f.GetQueue(key, maxSize)
	return queue.Enqueue(context.Background(), req)
}

func (f *QueueFactory) Dequeue(key string) (*QueuedRequest, error) {
	f.mu.RLock()
	queue, exists := f.queues[key]
	f.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	return queue.Dequeue(context.Background())
}

func (f *QueueFactory) GetStats(key string) map[string]interface{} {
	f.mu.RLock()
	queue, exists := f.queues[key]
	f.mu.RUnlock()

	if !exists {
		return map[string]interface{}{
			"error": "queue not found",
		}
	}

	return queue.GetStats()
}

func (f *QueueFactory) GetAllStats() map[string]map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for key, queue := range f.queues {
		stats[key] = queue.GetStats()
	}

	return stats
}

func (f *QueueFactory) ResetStats(key string) {
	f.mu.RLock()
	queue, exists := f.queues[key]
	f.mu.RUnlock()

	if exists {
		queue.ResetStats()
	}
}

func (f *QueueFactory) ResetAllStats() {
	f.mu.RLock()
	queues := make([]*PriorityQueue, 0, len(f.queues))
	for _, queue := range f.queues {
		queues = append(queues, queue)
	}
	f.mu.RUnlock()

	for _, queue := range queues {
		queue.ResetStats()
	}
}

func (f *QueueFactory) RemoveQueue(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.queues, key)
}

func (f *QueueFactory) GetQueueCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.queues)
}
