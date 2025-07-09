package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const POLLING_PERIOD time.Duration = time.Minute
const REDIS_JOBS_STREAM_KEY = "job_stream"

type Publisher struct {
	redis *redis.Client
	done  chan bool
}

func NewPublisher(redis *redis.Client) *Publisher {
	return &Publisher{
		redis: redis,
		done:  make(chan bool),
	}
}

func (p *Publisher) Start() {
	log.Println("PUBLISHER: Starting publisher...")

	currentUTCTime := time.Now().UTC()
	timeForNextInterval := currentUTCTime.Truncate(POLLING_PERIOD).Add(POLLING_PERIOD)
	sleepTime := time.Until(timeForNextInterval)

	go func() {
		time.Sleep(sleepTime)
		ticker := time.NewTicker(POLLING_PERIOD)
		// ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		p.publishToStream(timeForNextInterval)

		for {
			select {
			case <-p.done:
				return
			case <-ticker.C:
				// not using the time from ticker, because there can be delays/adds of milliseconds in ticker's time.
				// due to OS scheduler, CPU load or go runtime, ticker can fire slightly before or slightly after.
				// this drift in time causes mismatch of time from ticker and the exact time of job.
				// leading to mismatch i.e job's time is 04:07:00:0000 mismatches to ticker's time 04:06:59:9999
				// so we track our own next interval by adding a minute to the previous time and re-using it always
				timeForNextInterval = timeForNextInterval.Add(POLLING_PERIOD)
				// looping to publish, since we are fetching limited jobs from ZSET at a time.
				// So it will loop and fetch until there are no jobs
				for {
					hasMoreJobs, err := p.publishToStream(timeForNextInterval)
					if !hasMoreJobs || err != nil {
						break
					}
				}
			}
		}
	}()
}

func (p *Publisher) Stop() {
	log.Println("PUBLISHER: Stopping job publisher...")
	p.done <- true
	close(p.done)
}

func (p *Publisher) publishToStream(t time.Time) (bool, error) {
	log.Println("PUBLISHER: publishing for:", t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// a script to fetch and remove the members for each minute at redis server end
	// so that no multiple instance can have two same jobs
	// It works, as the redis server queues one script at a time coming from multiple instances.
	// So first instance's script will delete & return. second instance's script will get nothing
	script := `
		local key = KEYS[1]
		local min = ARGV[1]
		local max = ARGV[2]
		local limit = tonumber(ARGV[3])

		local members = redis.call("ZRANGEBYSCORE", key, min, max, "LIMIT", 0, limit)
		if #members > 0 then
			redis.call("ZREM", key, unpack(members))
		end
		return members
	`

	// fetching from ZSET
	jobScore := t.Unix()
	min := fmt.Sprintf("%d", jobScore)
	max := fmt.Sprintf("%d", jobScore)
	limit := 100
	res, err := p.redis.Eval(ctx, script, []string{REDIS_JOBS_ZSET_KEY}, min, max, limit).Result()
	if err != nil {
		log.Printf("PUBLISHER: error in executing lua script: %s", err.Error())
		return false, err
	}

	// checking if we have got results
	members, ok := res.([]any)
	if !ok {
		return false, fmt.Errorf("invalid results from script")
	}
	if len(members) == 0 {
		log.Printf("PUBLISHER: no jobs found for %d:%d at score: %d\n", t.Hour(), t.Minute(), jobScore)
		return false, nil
	}

	// publishing the jobs to redis stream
	pipe := p.redis.Pipeline()
	for _, member := range members {
		// fmt.Println(member)
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: REDIS_JOBS_STREAM_KEY,
			Values: map[string]any{
				"job": member,
			},
		})
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("PUBLISHER: Failed to publish job with pipeline: %s\n", err.Error())
	}

	hasMoreJobs := true
	if len(members) < limit {
		hasMoreJobs = false
	}
	return hasMoreJobs, nil
}
