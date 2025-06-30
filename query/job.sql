-- name: InsertJob :one
INSERT INTO job(state, description)
VALUES (@state::job_state, @args::jsonb)
RETURNING id;

-- name: GetJobs :many
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

-- name: CompleteJob :one
UPDATE job
SET state = 'completed'::job_state
WHERE id = @id::bigint
RETURNING *;

-- name: VacuumJobs :one
WITH deleted_jobs AS (
    DELETE FROM job
    WHERE id IN (
        SELECT id
        FROM job
        WHERE state = 'completed'
        ORDER BY id -- FIXME: maybe remove
        LIMIT @batch_size::integer
    )
    RETURNING *
)
SELECT count(*)
FROM deleted_jobs;
