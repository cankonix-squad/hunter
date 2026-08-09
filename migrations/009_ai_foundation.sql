BEGIN;

CREATE TABLE ai_prompt_versions (
    version text PRIMARY KEY,
    task text NOT NULL,
    schema_version text NOT NULL,
    system_prompt text NOT NULL,
    max_cost_micro bigint NOT NULL CHECK (max_cost_micro >= 0),
    active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ai_prompt_versions_one_active_idx ON ai_prompt_versions(task) WHERE active;

CREATE TABLE ai_results (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task text NOT NULL,
    prompt_version text NOT NULL REFERENCES ai_prompt_versions(version),
    input_fingerprint text NOT NULL,
    provider text NOT NULL,
    model text NOT NULL,
    schema_version text NOT NULL,
    result jsonb NOT NULL CHECK (jsonb_typeof(result) = 'object'),
    confidence_bp smallint NOT NULL CHECK (confidence_bp BETWEEN 0 AND 10000),
    evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'array'),
    fallback boolean NOT NULL DEFAULT false,
    input_tokens integer NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
    output_tokens integer NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),
    cost_micro bigint NOT NULL DEFAULT 0 CHECK (cost_micro >= 0),
    created_at timestamptz NOT NULL,
    UNIQUE (task, prompt_version, input_fingerprint, provider, model)
);

CREATE INDEX ai_results_cache_idx
    ON ai_results(task, prompt_version, input_fingerprint, provider, created_at DESC);

CREATE TABLE ai_daily_usage (
    provider text NOT NULL,
    usage_date date NOT NULL,
    requests integer NOT NULL DEFAULT 0 CHECK (requests >= 0),
    input_tokens bigint NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
    output_tokens bigint NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),
    cost_micro bigint NOT NULL DEFAULT 0 CHECK (cost_micro >= 0),
    PRIMARY KEY (provider, usage_date)
);

INSERT INTO ai_prompt_versions(version, task, schema_version, system_prompt, max_cost_micro, active)
VALUES ('classify-it/v1', 'classify_it', '1.0',
        'Classify only from supplied fields. Return strict JSON. Never infer missing qualifications.',
        25000, true);

INSERT INTO schema_migrations(version) VALUES (9);
COMMIT;
