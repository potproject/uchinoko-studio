package memory

import (
	"context"
	"log"
	"sync"
	"time"
)

var (
	workerOnce sync.Once
	stopOnce   sync.Once
	stopChan   chan struct{}
)

func StartWorker() {
	workerOnce.Do(func() {
		stopChan = make(chan struct{})
		if err := recoverRunningJobs(); err != nil {
			log.Printf("memory worker recover jobs error: %v", err)
		}
		go workerLoop()
	})
}

func StopWorker() {
	stopOnce.Do(func() {
		if stopChan != nil {
			close(stopChan)
		}
	})
}

func workerLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stopChan:
			return
		default:
		}

		job, ok, err := claimNextJob(context.Background())
		if err != nil {
			log.Printf("memory worker claim error: %v", err)
			select {
			case <-time.After(3 * time.Second):
			case <-stopChan:
				return
			}
			continue
		}
		if !ok {
			select {
			case <-ticker.C:
				continue
			case <-stopChan:
				return
			}
		}
		if err := processJob(context.Background(), job); err != nil {
			log.Printf("memory worker process %s error: %v", job.Type, err)
			if failErr := failJob(context.Background(), job.ID, job.Attempts, err); failErr != nil {
				log.Printf("memory worker failJob error: %v", failErr)
			}
			continue
		}
		if err := completeJob(context.Background(), job.ID); err != nil {
			log.Printf("memory worker completeJob error: %v", err)
		}
	}
}
