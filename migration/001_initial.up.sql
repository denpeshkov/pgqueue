CREATE TYPE job_state AS ENUM(
    'available',
    'running',
    'completed'
    -- 'failed'
);

-- TODO: Maybe UNLOGGED
CREATE TABLE IF NOT EXISTS job(
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY, -- TODO: Maybe ULID
    state job_state NOT NULL DEFAULT 'available'::job_state,
    description text
    --TODO: visibility_timeout timestampz NOT NULL DEFAULT NOW(), See: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html
);

CREATE OR REPLACE FUNCTION job_notify() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.state = 'available' THEN
        -- Same-payload notifications in a transaction are deduplicated.
        PERFORM pg_notify('job-channel', ''); -- No payload
    END IF;
    RETURN NULL;
END; $$
LANGUAGE plpgsql;

CREATE TRIGGER trigger_job_notify
AFTER INSERT ON job
FOR EACH ROW
EXECUTE PROCEDURE job_notify();
