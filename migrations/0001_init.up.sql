CREATE TABLE IF NOT EXISTS jobs(
    id SERIAL,
    hour INT NOT NULL,
    minute INT NOT NULL,
    type TEXT NOT NULL, --email, ping, slack, webhook
    payload JSONB NOT NULL,
    retries INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(hour, id)
) PARTITION BY RANGE(hour);

CREATE TABLE IF NOT EXISTS jobs_00_to_06 PARTITION OF jobs
    FOR VALUES FROM (0) TO (7);

CREATE TABLE IF NOT EXISTS jobs_07_to_12 PARTITION OF jobs
    FOR VALUES FROM (7) TO (13);

CREATE TABLE IF NOT EXISTS jobs_13_to_18 PARTITION OF jobs
    FOR VALUES FROM (13) TO (19);

CREATE TABLE IF NOT EXISTS jobs_19_to_23 PARTITION OF jobs
    FOR VALUES FROM (19) TO (24);


CREATE TABLE IF NOT EXISTS job_runs(
    id SERIAL PRIMARY KEY,
    job_id INT,
    status TEXT, --running, completed, failed, permanently_failed
    output TEXT,
    error TEXT,
    scheduled_at TIMESTAMP,
    completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);