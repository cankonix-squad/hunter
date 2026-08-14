package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var provisionableRoles = map[string]bool{
	"administrator": true, "operator": true, "business_development": true,
	"bid_manager": true, "technical_lead": true, "executive": true,
}

func (s *Service) ProvisionUser(ctx context.Context, email, name, password string, roles []string) (Principal, error) {
	email, name, roles, err := validateProvisionInput(email, name, password, roles)
	if err != nil {
		return Principal{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Principal{}, fmt.Errorf("hash provisioned password: %w", err)
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Principal{}, fmt.Errorf("begin user provisioning: %w", err)
	}
	defer tx.Rollback(ctx)
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO users(email,display_name,password_hash) VALUES ($1,$2,$3)
		RETURNING id::text`, email, name, string(hash)).Scan(&id)
	if err != nil {
		return Principal{}, fmt.Errorf("create user: %w", err)
	}
	for _, role := range roles {
		result, err := tx.Exec(ctx, `INSERT INTO user_roles(user_id,role_id)
			SELECT $1,id FROM roles WHERE name=$2`, id, role)
		if err != nil {
			return Principal{}, fmt.Errorf("assign role %s: %w", role, err)
		}
		if result.RowsAffected() != 1 {
			return Principal{}, fmt.Errorf("assign role %s: role is unavailable", role)
		}
	}
	metadata, _ := json.Marshal(map[string]any{"email": email, "roles": roles, "provisioner": "cli"})
	if _, err := tx.Exec(ctx, `INSERT INTO audit_logs(actor_id,action,object_type,object_id,metadata)
		VALUES (NULL,'identity.user_provisioned','user',$1,$2)`, id, metadata); err != nil {
		return Principal{}, fmt.Errorf("audit user provisioning: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Principal{}, fmt.Errorf("commit user provisioning: %w", err)
	}
	return Principal{UserID: id, Email: email, Name: name, Roles: roles}, nil
}

func validateProvisionInput(email, name, password string, roles []string) (string, string, []string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || name == "" {
		return "", "", nil, errors.New("valid email and non-empty display name are required")
	}
	if len(password) < 20 {
		return "", "", nil, errors.New("password must contain at least 20 characters")
	}
	seen := make(map[string]bool)
	cleanRoles := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" || seen[role] {
			continue
		}
		if !provisionableRoles[role] {
			return "", "", nil, fmt.Errorf("role %q is not provisionable", role)
		}
		seen[role] = true
		cleanRoles = append(cleanRoles, role)
	}
	sort.Strings(cleanRoles)
	return email, name, cleanRoles, nil
}
