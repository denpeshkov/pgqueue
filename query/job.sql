-- name: JobInsert :one
INSERT INTO job(state, description)
VALUES (@state::job_state, @args::jsonb)
RETURNING id;

-- name: JobGetAvailable :many
WITH available_jobs AS (
    SELECT id
    FROM job
    WHERE state = 'available'::job_state
    ORDER BY id ASC
    LIMIT @batch_size::integer
    FOR UPDATE SKIP LOCKED
)
UPDATE job
SET state = 'running'::job_state
FROM available_jobs
WHERE job.id = available_jobs.id
RETURNING job.*;