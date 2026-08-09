package opportunities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var validStatuses = map[string]bool{
	"announced": true, "registration_open": true, "registration_closed": true,
	"evaluation": true, "awarded": true, "contracted": true,
	"completed": true, "cancelled": true,
}

type Normalized struct {
	Title                string     `json:"title"`
	Agency               string     `json:"agency"`
	ProcurementType      string     `json:"procurement_type"`
	LifecycleStatus      string     `json:"lifecycle_status"`
	BudgetMinor          *int64     `json:"budget_minor,omitempty"`
	Currency             string     `json:"currency"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
}

func Normalize(metadata map[string]any) (Normalized, error) {
	result := Normalized{
		Title:           text(metadata, "title"),
		Agency:          text(metadata, "agency"),
		ProcurementType: text(metadata, "procurement_type"),
		LifecycleStatus: strings.ToLower(text(metadata, "status")),
		Currency:        strings.ToUpper(text(metadata, "currency")),
	}
	if result.Title == "" || result.Agency == "" || result.ProcurementType == "" {
		return Normalized{}, errors.New("title, agency, and procurement_type are required")
	}
	if !validStatuses[result.LifecycleStatus] {
		return Normalized{}, fmt.Errorf("unsupported lifecycle status %q", result.LifecycleStatus)
	}
	if len(result.Currency) != 3 {
		return Normalized{}, errors.New("currency must be a three-letter code")
	}
	if raw, ok := metadata["pagu_minor"]; ok {
		value, ok := integer(raw)
		if !ok || value < 0 {
			return Normalized{}, errors.New("pagu_minor must be a non-negative integer")
		}
		result.BudgetMinor = &value
	}
	if raw := text(metadata, "registration_deadline"); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return Normalized{}, errors.New("registration_deadline must be RFC3339")
		}
		value = value.UTC()
		result.RegistrationDeadline = &value
	}
	return result, nil
}

func text(metadata map[string]any, key string) string {
	value, _ := metadata[key].(string)
	return strings.TrimSpace(value)
}

func integer(value any) (int64, bool) {
	switch number := value.(type) {
	case int64:
		return number, true
	case int:
		return int64(number), true
	case float64:
		converted := int64(number)
		return converted, float64(converted) == number
	default:
		return 0, false
	}
}
