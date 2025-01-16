package utils

import "sync"

type UserRequestQueue struct {
	mu     sync.Mutex
	queues map[string]chan struct{}
}

func NewUserRequestQueue() *UserRequestQueue {
	return &UserRequestQueue{
		queues: make(map[string]chan struct{}),
	}
}

func (q *UserRequestQueue) getQueue(userID string) chan struct{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.queues[userID]; !exists {
		q.queues[userID] = make(chan struct{}, 1)
	}

	return q.queues[userID]
}

func (q *UserRequestQueue) Process(userID string, handler func()) {
	queue := q.getQueue(userID)

	// Enqueue the request
	queue <- struct{}{}

	go func() {
		// Process the request
		handler()

		// Dequeue the request
		<-queue
	}()
}
