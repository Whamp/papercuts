// Package papercuts implements capture, initialization, and managed guidance.
package papercuts

import "fmt"

// Severity classifies the impact of a papercut.
type Severity uint8

const (
	severityLow Severity = iota + 1
	severityMedium
	severityHigh
)

// SeverityDefinition contains the canonical wording for one severity.
type SeverityDefinition struct {
	Severity    Severity
	Value       string
	Meaning     string
	HelpSummary string
}

var severityDefinitions = [...]SeverityDefinition{
	{
		Severity:    severityLow,
		Value:       "low",
		Meaning:     "an avoidable detour that did not change your approach or confidence in the result.",
		HelpSummary: "Avoidable detour; approach and confidence remained intact",
	},
	{
		Severity:    severityMedium,
		Value:       "medium",
		Meaning:     "meaningful rework, repeated attempts, a workaround, a changed approach, or reduced confidence while the task remained safely completable.",
		HelpSummary: "Meaningful rework, retries, workaround, changed approach, or reduced confidence",
	},
	{
		Severity:    severityHigh,
		Value:       "high",
		Meaning:     "blocked completion, required human or environment intervention, or created a credible risk of an incorrect, destructive, or insecure result.",
		HelpSummary: "Blocked work, required intervention, or credible risk of a wrong, destructive, or insecure result",
	},
}

// ParseSeverity parses one canonical lowercase severity value.
func ParseSeverity(value string) (Severity, error) {
	for _, definition := range severityDefinitions {
		if definition.Value == value {
			return definition.Severity, nil
		}
	}
	return 0, fmt.Errorf("invalid severity %q", value)
}

// SeverityDefinitions returns the canonical severities in ascending order.
func SeverityDefinitions() []SeverityDefinition {
	definitions := make([]SeverityDefinition, len(severityDefinitions))
	copy(definitions, severityDefinitions[:])
	return definitions
}

// String returns the canonical lowercase severity value.
func (s Severity) String() string {
	for _, definition := range severityDefinitions {
		if definition.Severity == s {
			return definition.Value
		}
	}
	return ""
}
