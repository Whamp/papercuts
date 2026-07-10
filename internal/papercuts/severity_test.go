package papercuts

import "testing"

func TestParseSeverityAcceptsCanonicalValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"low", "medium", "high"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			got, err := ParseSeverity(value)
			if err != nil {
				t.Errorf("ParseSeverity(%q) error = %v, want nil", value, err)
			}
			if got.String() != value {
				t.Errorf("ParseSeverity(%q).String() = %q, want %q", value, got.String(), value)
			}
		})
	}
}

func TestParseSeverityRejectsEveryOtherValue(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "LOW", "Medium", "critical", "1", " low"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseSeverity(value); err == nil {
				t.Errorf("ParseSeverity(%q) returned nil error, want rejection", value)
			}
		})
	}
}
