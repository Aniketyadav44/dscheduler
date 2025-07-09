import random
import json
from datetime import datetime, timedelta

# -------------------------
# 🔧 Configurable Variables
# -------------------------

hour = 6                    # UTC hour
minute_start = 1
minute_end = 10             # inclusive

heavy_minutes = ["06:01", "06:05", "06:07"]
heavy_jobs_count = 20000

min_jobs = 1000
max_jobs = 5000

output_file = "insert_jobs.sql"
batch_size = 1000

valid_urls = [
    "https://anikety.com",
    "https://openai.com",
    "https://example.com",
    "https://google.com",
    "https://github.com",
]

invalid_urls = [
    "http://fail.test",
    "http://localhost:1234",
    "http://127.0.0.1:9999",
    "http://bad.domain.xyz",
    "http://not.reachable",
]

# -------------------------
# 🔁 Job Generation Logic
# -------------------------

def generate_jobs_for_minute(hour, minute, count):
    rows = []
    utc_time = datetime(2024, 1, 1, hour, minute)
    ist_time = utc_time + timedelta(hours=5, minutes=30)
    ist_hour = ist_time.hour
    ist_minute = ist_time.minute
    
    invalid_probability = random.uniform(0.01, 0.05)  # Between 1% and 5%

    for _ in range(count):
        url = random.choice(invalid_urls if random.random() < invalid_probability else valid_urls)
        payload = {
            "url": url,
            "hour": f"{ist_hour:02}",
            "minute": f"{ist_minute:02}"
        }
        row = f"({hour}, {minute}, 'ping', '{json.dumps(payload)}')"
        rows.append(row)

    return rows

# -------------------------
# 🛠️ Create SQL Insert File
# -------------------------

sql_lines = []

# Step 1: Range-based job generation
for minute in range(minute_start, minute_end + 1):
    job_count = random.randint(min_jobs, max_jobs)
    job_rows = generate_jobs_for_minute(hour, minute, job_count)

    for i in range(0, len(job_rows), batch_size):
        batch = job_rows[i:i+batch_size]
        sql = f"INSERT INTO jobs(hour, minute, type, payload)\nVALUES\n" + ",\n".join(batch) + ";\n"
        sql_lines.append(sql)

# Step 2: Heavy fixed-minute jobs
for hm in heavy_minutes:
    h, m = map(int, hm.split(":"))
    if h == hour and minute_start <= m <= minute_end:
        job_rows = generate_jobs_for_minute(h, m, heavy_jobs_count)

    for i in range(0, len(job_rows), batch_size):
        batch = job_rows[i:i+batch_size]
        sql = f"INSERT INTO jobs(hour, minute, type, payload)\nVALUES\n" + ",\n".join(batch) + ";\n"
        sql_lines.append(sql)

# Write to file
with open(output_file, "w") as f:
    f.writelines(sql_lines)

print(f"✅ SQL file '{output_file}' with job inserts generated.")
