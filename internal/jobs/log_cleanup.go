package jobs

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/go-links/internal/models"
)

type LogCleanupJob struct {
	logManager *models.LogManager
	interval   time.Duration
	stopChan   chan struct{}
}

func NewLogCleanupJob(logManager *models.LogManager, interval time.Duration) *LogCleanupJob {
	return &LogCleanupJob{
		logManager: logManager,
		interval:   interval,
		stopChan:   make(chan struct{}),
	}
}

func (j *LogCleanupJob) Start() {
	ticker := time.NewTicker(j.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
				if err := j.logManager.CleanupOldLogs(ctx); err != nil {
					log.Printf("Error cleaning up logs: %v", err)
				}
				cancel()
			case <-j.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (j *LogCleanupJob) Stop() {
	close(j.stopChan)
} 