package workerpool

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type WorkerPool struct {
	pool       chan func()
	closed     bool
	wg         sync.WaitGroup
	workerMux  sync.Mutex
	workersNum int
}

func NewPool(size int) *WorkerPool {
	if size < 1 {
		size = 1
	}

	pool := &WorkerPool{
		pool: make(chan func()),
	}

	pool.addWorkers(size)

	return pool
}

func (p *WorkerPool) Put(job func()) {
	if p.closed {
		return
	}

	p.pool <- job
}

func (p *WorkerPool) Close() {
	if p.closed {
		return
	}

	p.closed = true
	close(p.pool)
	p.wg.Wait()
}

func (p *WorkerPool) AddWorkers(n int) {
	if p.closed || n <= 0 {
		return
	}

	p.addWorkers(n)
}

func (p *WorkerPool) Remove(n int) {
	if p.closed || n <= 0 {
		return
	}

	p.workerMux.Lock()
	defer p.workerMux.Unlock()

	if n >= p.workersNum {
		n = p.workersNum - 1
		if n < 1 {
			n = 1
		}
	}

	for range n {
		p.Put(func() {})

		p.workersNum--
	}
}

func (p *WorkerPool) Count() int {
	p.workerMux.Lock()
	defer p.workerMux.Unlock()

	return p.workersNum
}

func (p *WorkerPool) addWorkers(n int) {
	p.workerMux.Lock()
	defer p.workerMux.Unlock()

	for range n {
		p.wg.Add(1)

		p.workersNum++

		workerID := p.workersNum

		go func() {
			defer p.wg.Done()

			for job := range p.pool {
				log.Info().Msgf("worker %d processing job", workerID)
				job()
			}
		}()
	}
}
