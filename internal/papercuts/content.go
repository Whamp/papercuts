package papercuts

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// ValidationError identifies invalid capture input.
type ValidationError struct {
	Field  string
	Reason string
}

// Error returns a field-specific validation error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Reason)
}

func prepareEntry(
	severity Severity,
	description string,
	reporter *string,
	model *string,
	now func() time.Time,
) (entry, error) {
	if severity < severityLow || severity > severityHigh {
		return entry{}, &ValidationError{Field: "severity", Reason: "is invalid"}
	}
	if !utf8.ValidString(description) {
		return entry{}, &ValidationError{Field: "description", Reason: "must be valid UTF-8"}
	}
	description = strings.TrimSpace(description)
	if description == "" {
		return entry{}, &ValidationError{Field: "description", Reason: "is empty after trimming"}
	}

	validatedReporter, err := validateLabel("reporter", reporter)
	if err != nil {
		return entry{}, err
	}
	validatedModel, err := validateLabel("model", model)
	if err != nil {
		return entry{}, err
	}

	return entry{
		capturedAt:  now().UTC(),
		severity:    severity,
		description: description,
		reporter:    validatedReporter,
		model:       validatedModel,
	}, nil
}

func validateLabel(field string, value *string) (optionalLabel, error) {
	if value == nil {
		return optionalLabel{}, nil
	}
	if !utf8.ValidString(*value) {
		return optionalLabel{}, &ValidationError{Field: field, Reason: "must be valid UTF-8"}
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return optionalLabel{}, &ValidationError{Field: field, Reason: "must not be empty"}
	}
	if strings.ContainsAny(trimmed, "\x00\r\n") {
		return optionalLabel{}, &ValidationError{Field: field, Reason: "must not contain NUL or a line break"}
	}
	return optionalLabel{present: true, value: trimmed}, nil
}
