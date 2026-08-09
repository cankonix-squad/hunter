package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ db *pgxpool.Pool }

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) Get(ctx context.Context, key CacheKey) (Envelope, bool, error) {
	var value Envelope
	var resultJSON, evidenceJSON []byte
	err := s.db.QueryRow(ctx, `SELECT schema_version,result,confidence_bp,evidence,provider,model,
		prompt_version,input_fingerprint,created_at,fallback FROM ai_results
		WHERE task=$1 AND prompt_version=$2 AND input_fingerprint=$3 AND provider=$4
		ORDER BY created_at DESC LIMIT 1`, key.Task, key.PromptVersion, key.InputFingerprint, key.Provider).
		Scan(&value.SchemaVersion, &resultJSON, &value.ConfidenceBP, &evidenceJSON, &value.Provider,
			&value.Model, &value.PromptVersion, &value.InputFingerprint, &value.CreatedAt, &value.Fallback)
	if err == pgx.ErrNoRows {
		return Envelope{}, false, nil
	}
	if err != nil {
		return Envelope{}, false, err
	}
	value.Task = key.Task
	if err := json.Unmarshal(resultJSON, &value.Result); err != nil {
		return Envelope{}, false, fmt.Errorf("decode cached AI result: %w", err)
	}
	if err := json.Unmarshal(evidenceJSON, &value.Evidence); err != nil {
		return Envelope{}, false, fmt.Errorf("decode cached AI evidence: %w", err)
	}
	return value, true, nil
}

func (s *PostgresStore) Put(ctx context.Context, key CacheKey, value Envelope, response GenerateResponse) error {
	resultJSON, err := json.Marshal(value.Result)
	if err != nil {
		return err
	}
	evidenceJSON, err := json.Marshal(value.Evidence)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `INSERT INTO ai_results(task,prompt_version,input_fingerprint,provider,model,
		schema_version,result,confidence_bp,evidence,fallback,input_tokens,output_tokens,cost_micro,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (task,prompt_version,input_fingerprint,provider,model) DO NOTHING`,
		key.Task, key.PromptVersion, key.InputFingerprint, key.Provider, key.Model, value.SchemaVersion,
		resultJSON, value.ConfidenceBP, evidenceJSON, value.Fallback, response.InputTok, response.OutputTok,
		response.CostMicro, value.CreatedAt)
	return err
}

type PostgresBudget struct {
	db          *pgxpool.Pool
	dailyLimit  int64
	perRunLimit int64
	now         func() time.Time
}

func NewPostgresBudget(db *pgxpool.Pool, perRunLimit, dailyLimit int64) *PostgresBudget {
	return &PostgresBudget{db: db, perRunLimit: perRunLimit, dailyLimit: dailyLimit, now: time.Now}
}

func (b *PostgresBudget) Allow(ctx context.Context, provider string, maximum int64) (bool, error) {
	if b.perRunLimit > 0 && maximum > b.perRunLimit {
		return false, nil
	}
	if b.dailyLimit <= 0 {
		return true, nil
	}
	var spent int64
	err := b.db.QueryRow(ctx, `SELECT coalesce(cost_micro,0) FROM ai_daily_usage
		WHERE provider=$1 AND usage_date=$2`, provider, b.now().UTC().Format("2006-01-02")).Scan(&spent)
	if err == pgx.ErrNoRows {
		spent = 0
	} else if err != nil {
		return false, err
	}
	return spent+maximum <= b.dailyLimit, nil
}

func (b *PostgresBudget) Record(ctx context.Context, provider string, response GenerateResponse) error {
	_, err := b.db.Exec(ctx, `INSERT INTO ai_daily_usage(provider,usage_date,requests,input_tokens,output_tokens,cost_micro)
		VALUES ($1,$2,1,$3,$4,$5) ON CONFLICT (provider,usage_date) DO UPDATE SET
		requests=ai_daily_usage.requests+1,input_tokens=ai_daily_usage.input_tokens+excluded.input_tokens,
		output_tokens=ai_daily_usage.output_tokens+excluded.output_tokens,
		cost_micro=ai_daily_usage.cost_micro+excluded.cost_micro`, provider, b.now().UTC().Format("2006-01-02"),
		response.InputTok, response.OutputTok, response.CostMicro)
	return err
}
