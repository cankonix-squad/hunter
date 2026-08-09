package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type Principal struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
}

type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Principal    Principal `json:"user"`
}

type claims struct {
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type Service struct {
	db         *pgxpool.Pool
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(db *pgxpool.Pool, secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{db: db, secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *Service) BootstrapAdmin(ctx context.Context, email, name, password string) error {
	if email == "" && password == "" {
		return nil
	}
	if email == "" || len(password) < 12 {
		return fmt.Errorf("bootstrap administrator needs an email and password of at least 12 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}
	_, err = s.db.Exec(ctx, `
		WITH inserted AS (
			INSERT INTO users(email, display_name, password_hash)
			VALUES (lower($1), $2, $3)
			ON CONFLICT (email) DO NOTHING
			RETURNING id
		)
		INSERT INTO user_roles(user_id, role_id)
		SELECT inserted.id, roles.id FROM inserted CROSS JOIN roles WHERE roles.name = 'administrator'`, email, name, string(hash))
	if err != nil {
		return fmt.Errorf("bootstrap administrator: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (Tokens, error) {
	var p Principal
	var passwordHash string
	err := s.db.QueryRow(ctx, `
		SELECT u.id::text, u.email, u.display_name, u.password_hash,
		       ARRAY(SELECT r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = u.id ORDER BY r.name)
		FROM users u
		WHERE u.email = lower($1) AND u.active
		`, strings.TrimSpace(email)).Scan(&p.UserID, &p.Email, &p.Name, &passwordHash, &p.Roles)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil) {
		return Tokens{}, ErrInvalidCredentials
	}
	if err != nil {
		return Tokens{}, fmt.Errorf("load user: %w", err)
	}
	return s.issue(ctx, p, "")
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (Tokens, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Tokens{}, fmt.Errorf("begin refresh: %w", err)
	}
	defer tx.Rollback(ctx)

	var sessionID, familyID string
	var p Principal
	var rotatedAt, revokedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT s.id::text, s.family_id::text, s.rotated_at, s.revoked_at,
		       u.id::text, u.email, u.display_name,
		       ARRAY(SELECT r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = u.id ORDER BY r.name)
		FROM refresh_sessions s
		JOIN users u ON u.id = s.user_id AND u.active
		WHERE s.token_hash = $1 AND s.expires_at > now()
		FOR UPDATE OF s`, hash[:]).Scan(&sessionID, &familyID, &rotatedAt, &revokedAt, &p.UserID, &p.Email, &p.Name, &p.Roles)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tokens{}, ErrInvalidRefreshToken
	}
	if err != nil {
		return Tokens{}, fmt.Errorf("load refresh session: %w", err)
	}
	if revokedAt != nil || rotatedAt != nil {
		_, _ = tx.Exec(ctx, `UPDATE refresh_sessions SET revoked_at = now() WHERE family_id = $1 AND revoked_at IS NULL`, familyID)
		_ = tx.Commit(ctx)
		return Tokens{}, ErrInvalidRefreshToken
	}

	raw, newHash, err := newOpaqueToken()
	if err != nil {
		return Tokens{}, err
	}
	var newID string
	err = tx.QueryRow(ctx, `INSERT INTO refresh_sessions(family_id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, now() + $4::interval) RETURNING id::text`, familyID, p.UserID, newHash, interval(s.refreshTTL)).Scan(&newID)
	if err != nil {
		return Tokens{}, fmt.Errorf("rotate refresh session: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE refresh_sessions SET rotated_at = now(), replaced_by = $1 WHERE id = $2`, newID, sessionID); err != nil {
		return Tokens{}, fmt.Errorf("close refresh session: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Tokens{}, fmt.Errorf("commit refresh: %w", err)
	}
	return s.accessTokens(p, raw)
}

func (s *Service) Revoke(ctx context.Context, rawToken string) error {
	hash := sha256.Sum256([]byte(rawToken))
	_, err := s.db.Exec(ctx, `UPDATE refresh_sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`, hash[:])
	return err
}

func (s *Service) ParseAccess(raw string) (Principal, error) {
	token, err := jwt.ParseWithClaims(raw, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	}, jwt.WithAudience("cankonix-oi-web"), jwt.WithIssuer("cankonix-oi"), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return Principal{}, ErrInvalidCredentials
	}
	c, ok := token.Claims.(*claims)
	if !ok {
		return Principal{}, ErrInvalidCredentials
	}
	return Principal{UserID: c.Subject, Email: c.Email, Name: c.Name, Roles: c.Roles}, nil
}

func (s *Service) issue(ctx context.Context, p Principal, familyID string) (Tokens, error) {
	raw, hash, err := newOpaqueToken()
	if err != nil {
		return Tokens{}, err
	}
	query := `INSERT INTO refresh_sessions(family_id, user_id, token_hash, expires_at)
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, $3, now() + $4::interval)`
	if _, err = s.db.Exec(ctx, query, familyID, p.UserID, hash, interval(s.refreshTTL)); err != nil {
		return Tokens{}, fmt.Errorf("create refresh session: %w", err)
	}
	return s.accessTokens(p, raw)
}

func (s *Service) accessTokens(p Principal, refresh string) (Tokens, error) {
	now := time.Now().UTC()
	expires := now.Add(s.accessTTL)
	c := claims{Email: p.Email, Name: p.Name, Roles: p.Roles, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: "cankonix-oi", Subject: p.UserID, Audience: jwt.ClaimStrings{"cankonix-oi-web"},
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expires),
	}}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.secret)
	if err != nil {
		return Tokens{}, fmt.Errorf("sign access token: %w", err)
	}
	return Tokens{AccessToken: signed, RefreshToken: refresh, TokenType: "Bearer", ExpiresAt: expires, Principal: p}, nil
}

func newOpaqueToken() (string, []byte, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	hash := sha256.Sum256([]byte(raw))
	return raw, hash[:], nil
}

func interval(d time.Duration) string { return fmt.Sprintf("%f seconds", d.Seconds()) }
