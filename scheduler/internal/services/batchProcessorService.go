package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Aniketyadav44/dscheduler/scheduler/internal/models"
	"github.com/Aniketyadav44/dscheduler/scheduler/internal/utils"
	"github.com/redis/go-redis/v9"
)

const BATCH_PERIOD time.Duration = 10 * time.Minute
const MAX_BATCH_RETIRES int = 3
const REDIS_JOBS_ZSET_KEY string = "jobs"

type BatchProcessor struct {
	db    *sql.DB
	redis *redis.Client
	done  chan bool // for stopping the ticker
}

func NewBatchProcessor(db *sql.DB, redis *redis.Client) *BatchProcessor {
	return &BatchProcessor{
		db:    db,
		redis: redis,
		done:  make(chan bool),
	}
}

// This starts the processor ticker in a goroutine
func (b *BatchProcessor) Start() {
	log.Println("BATCH_PROCESSOR: Starting batch processor...")
	// calculating & sleeping until next 10 rounded time
	currentUTCTime := time.Now().UTC()
	next10Interval := currentUTCTime.Truncate(BATCH_PERIOD).Add(BATCH_PERIOD)
	// next10IntervalAt := currentUTCTime.Truncate(time.Minute).Add(time.Minute)
	sleepMins := time.Until(next10Interval)

	go func() {
		time.Sleep(sleepMins)
		ticker := time.NewTicker(BATCH_PERIOD)
		// ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		b.processJobBatch(next10Interval)

		for {
			select {
			case <-b.done:
				return
			case <-ticker.C:
				next10Interval = next10Interval.Add(BATCH_PERIOD)
				// will retry if getting error with 500ms backoff
				for i := 1; i <= MAX_BATCH_RETIRES; i++ {
					err := b.processJobBatch(next10Interval)
					if err == nil {
						break
					}
					log.Printf("BATCH_PROCESSOR: Batch run no. %d failed: %s\n", i, err.Error())
					time.Sleep(500 * time.Millisecond)
				}
			}
		}
	}()
}

// for stopping the batch processor ticker and return from goroutine
func (b *BatchProcessor) Stop() {
	log.Println("BATCH_PROCESSOR: Stopping batch processor...")
	b.done <- true
}

// Pull job from db and pushes to ZSET with lock check. 3 retries on error
func (b *BatchProcessor) processJobBatch(t time.Time) error {
	start := t
	end := start.Add(BATCH_PERIOD)
	log.Printf("BATCH_PROCESSOR: Initiating batch process for start: %s to end: %s\n", start, end)

	// trying to get the redis lock
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	batchLockKey := fmt.Sprintf("batch_lock_%d_%d", start.Minute(), end.Minute())
	hostname, _ := os.Hostname()
	batchLockValue := fmt.Sprintf("%s|%d", hostname, os.Getpid())
	ok, err := b.redis.SetNX(ctx, batchLockKey, batchLockValue, 2*time.Minute).Result()
	if !ok {
		// this batch is locked by some other scheduler instance.
		log.Printf("BATCH_PROCESSOR: Already locked: cannot lock batch for %s to %s\n", start, end)
		return nil
	}
	if err != nil {
		return err
	}
	log.Println("BATCH_PROCESSOR: Got redis lock for: ", batchLockKey)

	// pulling job batch from database
	jobs, err := b.pullFromDB(start, end)
	if err != nil {
		utils.ReleaseRedisLock(b.redis, batchLockKey, batchLockValue)
		return err
	}

	// if no jobs, then returing
	if len(jobs) == 0 {
		log.Println("BATCH_PROCESSOR: No jobs found for the time range, returning!")
		utils.ReleaseRedisLock(b.redis, batchLockKey, batchLockValue)
		return nil
	}

	// pushing the job batch to zset
	err = b.pushToRedisSet(jobs)
	if err != nil {
		utils.ReleaseRedisLock(b.redis, batchLockKey, batchLockValue)
		return err
	}

	// cleaning past jobs from ZSET
	timePoint := int64(start.Unix())
	b.clearPastJobs(timePoint)

	// finally releasing lock on success completion
	utils.ReleaseRedisLock(b.redis, batchLockKey, batchLockValue)

	return err
}

// Pull jobs from db for the time range
func (b *BatchProcessor) pullFromDB(start, end time.Time) ([]*models.Job, error) {
	log.Printf("BATCH_PROCESSOR: Pulling from database from %d:%d to %d:%d", start.Hour(), start.Minute()+1, end.Hour(), end.Minute()+1)
	query := `
		SELECT id, hour, minute, type, payload
		FROM jobs
		WHERE (hour > $1 OR (hour = $1 AND minute >= $2)) AND (hour < $3 OR (hour = $3 AND minute < $4))
	`

	// doing start.Minute+1 and end.Minute+1 to avoid overlapping times.
	// i.e for >=12:10 and <12:20 gives time 12:10 to 12:19
	// but adding 1 minute will give 12:11 to 12:20
	// so when next batch processes at 12:20, the job at 12:20 wont't be missed which will come in next batch
	rows, err := b.db.Query(query, start.Hour(), start.Minute()+1, end.Hour(), end.Minute()+1)
	if err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0)
	for rows.Next() {
		var job models.Job
		var jobPayload string
		if err := rows.Scan(&job.Id, &job.Hour, &job.Minute, &job.Type, &jobPayload); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(jobPayload), &job.Payload)
		// the start will always be at 00,10,20... that's why differencing it from the job's minute
		job.ScheduledTime = start.Add(time.Duration(job.Minute-start.Minute()) * time.Minute)
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// Pushes jobs to ZSET
func (b *BatchProcessor) pushToRedisSet(jobs []*models.Job) error {
	log.Println("BATCH_PROCESSOR: Pushing to redis zset")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	jobZs := make([]redis.Z, 0)
	for _, job := range jobs {
		jobData, err := json.Marshal(job)
		if err != nil {
			continue
		}
		jobScore := float64(job.ScheduledTime.Unix())
		jobZs = append(jobZs, redis.Z{
			Score:  jobScore,
			Member: jobData,
		})
	}
	// adding all jobs to the "jobs" sorted set
	err := b.redis.ZAdd(ctx, REDIS_JOBS_ZSET_KEY, jobZs...).Err()
	return err
}

// Delete past jobs in ZSET before the start time i.e 12:00/12:10/12:20..
func (b *BatchProcessor) clearPastJobs(timePoint int64) {
	log.Printf("BATCH_PROCESSOR: Clearing jobs in ZSET with score less than %d\n", timePoint)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := b.redis.ZRemRangeByScore(ctx, REDIS_JOBS_ZSET_KEY, "-inf", fmt.Sprintf("(%d", timePoint)).Result()
	if err != nil {
		log.Println("BATCH_PROCESSOR: Error in clearing past jobs: ", err.Error())
	}
}
