package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Aniketyadav44/dscheduler/worker/internal/models"
	"github.com/redis/go-redis/v9"
)

const REDIS_JOBS_STREAM_KEY string = "job_stream"
const REDIS_STREAM_GROUP_KEY string = "job_stream_group"
const MAX_JOB_RETRIES int = 3
const WORKERPOOL_COUNT int = 10

// job statuses
const (
	JOB_COMPLETED          = "completed"
	JOB_FAILED             = "failed"
	JOB_PERMANENTLY_FAILED = "permanently_failed"
)

type Consumer struct {
	dbService      *DBService
	redis          *redis.Client
	name           string
	done           chan bool
	workerPoolChan chan *redis.XMessage
}

func NewConsumer(dbService *DBService, r *redis.Client) *Consumer {
	hostname, _ := os.Hostname()
	consumerName := fmt.Sprintf("%s|%d", hostname, os.Getpid())
	return &Consumer{
		dbService:      dbService,
		redis:          r,
		name:           consumerName,
		done:           make(chan bool),
		workerPoolChan: make(chan *redis.XMessage),
	}
}

// first checks/creates stream & jobs and then start go routing to keep getting stream data
func (c *Consumer) Start() {
	log.Printf("CONSUMER: [%s] starting consumer service...", c.name)

	// checking for stream and consumer group
	if err := c.checkAndCreateConsumerGroup(); err != nil {
		log.Printf("CONSUMER: [%s] error in stream & group check: %s", c.name, err.Error())
		return
	}

	go func() {
		for {
			select {
			case <-c.done:
				log.Printf("CONSUMER: [%s] exiting redis consumer goroutine", c.name)
				return
			default:
				// adding this instance to the redis stream's consumer group
				res, err := c.redis.XReadGroup(context.Background(), &redis.XReadGroupArgs{
					Streams:  []string{REDIS_JOBS_STREAM_KEY, ">"},
					Group:    REDIS_STREAM_GROUP_KEY,
					Consumer: c.name,
					Count:    1000,
					Block:    5 * time.Second,
				}).Result()
				if err != nil && err != redis.Nil {
					log.Printf("CONSUMER: [%s] reading from stream group error: %s", c.name, err.Error())
					continue
				}

				for _, stream := range res {
					for _, msg := range stream.Messages {
						c.workerPoolChan <- &msg
					}
				}
			}
		}
	}()
	// why 20 count and 5 second block?
	// we have 5 second timeouts for each job execution with their retry & status writes to jobs & job runs table
	// so consider upper case, 20 jobs can take 100s on each instance

	// starting a worker pool
	for i := 1; i <= WORKERPOOL_COUNT; i++ {
		go func(id int) {
			log.Printf("CONSUMER: [%s] WORKERPOOL - starting goroutine: %d", c.name, id)
			for jobMsg := range c.workerPoolChan {
				c.processJob(jobMsg, id)
			}
		}(i)
	}
}

func (c *Consumer) Stop() {
	log.Printf("CONSUMER: [%s] stopping consumer service...", c.name)
	c.done <- true
	close(c.done)
	close(c.workerPoolChan)

	// deleting this consumer from the consumer group
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := c.redis.XGroupDelConsumer(ctx, REDIS_JOBS_STREAM_KEY, REDIS_STREAM_GROUP_KEY, c.name).Err()
	if err != nil {
		log.Printf("CONSUMER: [%s] error in deleting consumer: %s", c.name, err.Error())
	} else {
		log.Printf("CONSUMER: [%s] deleted consumer from consumer group: %s", c.name, REDIS_STREAM_GROUP_KEY)
	}
}

// this checks if stream & consumer group already exists. if not, create them
func (c *Consumer) checkAndCreateConsumerGroup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.redis.XGroupCreateMkStream(ctx, REDIS_JOBS_STREAM_KEY, REDIS_STREAM_GROUP_KEY, "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

// processing a job
func (c *Consumer) processJob(msg *redis.XMessage, jobId int) {
	jobStr := msg.Values["job"].(string)
	fmt.Printf("CONSUMER: [%s] processing job: msgID: %s at worker goroutine: %d\n", c.name, msg.ID, jobId)

	var job models.Job
	if err := json.Unmarshal([]byte(jobStr), &job); err != nil {
		log.Printf("CONSUMER: [%s] error in parsing job message: %s", c.name, err.Error())
		return
	}

	var jobErr error
	var output string
	switch job.Type {
	case "ping":
		output, jobErr = processPingJob(&job)
	case "email":
		output, jobErr = processEmailJob(&job)
	case "slack":
		output, jobErr = processSlackJob(&job)
	case "webhook":
		output, jobErr = processWebhookJob(&job)
	}

	c.handleJobResults(msg, &job, output, jobErr)
}

// handle job error
func (c *Consumer) handleJobResults(msg *redis.XMessage, job *models.Job, output string, err error) {
	jobRunEntry := &models.JobRun{
		JobId:       job.Id,
		ScheduledAt: job.ScheduledTime,
	}
	if err != nil {
		// job failed
		jobRunEntry.Error = sql.NullString{
			String: err.Error(),
			Valid:  true,
		}

		jobRetries, err := c.dbService.getJobRetryCount(job.Id, job.Hour)
		if err != nil {
			return
		}

		if jobRetries >= MAX_JOB_RETRIES {
			// job permanently failed
			jobRunEntry.Status = JOB_PERMANENTLY_FAILED
		} else {
			jobRunEntry.Status = JOB_FAILED
		}
		c.dbService.registerJobEntry(job, jobRunEntry)

		c.removeJobFromStream(msg)
		// adding the failed job back to stream for retries
		if jobRunEntry.Status == JOB_FAILED {
			time.Sleep(300 * time.Millisecond)
			jobPayload, _ := json.Marshal(job)
			err := c.redis.XAdd(context.Background(), &redis.XAddArgs{
				Stream: REDIS_JOBS_STREAM_KEY,
				Values: map[string]any{
					"job": string(jobPayload),
				},
			}).Err()
			if err != nil {
				log.Printf("CONSUMER: [%s] cannot requeue job to stream: %s", c.name, err.Error())
			}
		}
	} else {
		// job completed
		jobRunEntry.Status = JOB_COMPLETED
		jobRunEntry.Output = sql.NullString{
			String: output,
			Valid:  true,
		}
		c.dbService.registerJobEntry(job, jobRunEntry)
		c.removeJobFromStream(msg)
	}
}

// acknowledging msg and removing it from the stream
func (c *Consumer) removeJobFromStream(msg *redis.XMessage) {
	c.redis.XAck(context.Background(), REDIS_JOBS_STREAM_KEY, REDIS_STREAM_GROUP_KEY, msg.ID)
	c.redis.XDel(context.Background(), REDIS_JOBS_STREAM_KEY, msg.ID)
}
