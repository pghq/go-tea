// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package worker provides a background worker for offline processing.
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pghq/go-museum/museum/errors"
	"github.com/pghq/go-museum/museum/log"
)

const (
	// DefaultWorkerInstances is the default number of simultaneous workers
	DefaultWorkerInstances   = 1

	// DefaultWorkerInterval is the default period between running batches of jobs
	DefaultWorkerInterval    = 5 * time.Second
)

// Worker is an instance of a background worker.
type Worker struct{
	instances  int
	interval   time.Duration
	stopped    chan struct{}
	wg sync.WaitGroup
}

// New provides a new worker instance.
func New() *Worker {
	w := Worker{
		instances: DefaultWorkerInstances,
		interval:  DefaultWorkerInterval,
		stopped:   make(chan struct{}, 1),
	}

	return &w
}

func (w *Worker) Start(jobs ...func(context.Context)){
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < w.instances; i++ {
		w.wg.Add(1)
		go w.start(ctx, i+1, jobs)
	}

	log.Info(fmt.Sprintf("worker: workers=%d, started", w.instances))
	<-w.stopped
	w.wg.Wait()
	log.Info("worker: stopped")
}

func (w *Worker) Concurrent(instances uint) *Worker {
	w.instances = int(instances)

	return w
}

func (w *Worker) At(interval time.Duration) *Worker {
	w.interval = interval

	return w
}

func (w *Worker) Stop(){
	w.stopped <- struct{}{}
}

// do is the unit of work done for a single instance of the worker.
func (w *Worker) start(ctx context.Context, instance int, jobs []func(context.Context)) {
	defer w.wg.Done()

	stopped := make(chan struct{})
	ticker := time.NewTicker(w.interval)
	go func() {
		for {
			<-ticker.C
			if len(w.stopped) > 0{
				stopped <- struct{}{}
				break
			}

			for i, job := range jobs {
				go func(i int, job func(context.Context)) {
					defer func(){
						if err := recover(); err != nil{
							errors.Recover(err)
						}
					}()

					log.Debug(fmt.Sprintf("worker: instance=%d, job=%d, started", instance, i))
					ctx, cancel := context.WithTimeout(ctx, w.interval)
					job(ctx)
					cancel()
					log.Debug(fmt.Sprintf("worker: instance=%d, job=%d, finished", instance, i))
				}(i, job)
			}
		}
	}()

	log.Info(fmt.Sprintf("worker: instance=%d, started", instance))
	select{
	case <-stopped:
	case <-ctx.Done():
	}
	log.Info(fmt.Sprintf("worker: instance=%d, stopped", instance))
}
