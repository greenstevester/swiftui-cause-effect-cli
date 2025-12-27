package suggestions

import (
	"testing"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/issues"
)

func TestGenerateFixes_ExcessiveRerender(t *testing.T) {
	issue := issues.Issue{
		Type:     issues.IssueExcessiveRerender,
		Severity: issues.SeverityHigh,
		Title:    "Excessive re-renders in ItemRow",
	}

	fixes := GenerateFixes(issue)
	if len(fixes) == 0 {
		t.Error("Expected fixes for excessive rerender issue")
	}

	// Check that fixes have required fields
	for _, fix := range fixes {
		if fix.ID == "" {
			t.Error("Fix missing ID")
		}
		if fix.Approach == "" {
			t.Error("Fix missing Approach")
		}
		if fix.Description == "" {
			t.Error("Fix missing Description")
		}
		if fix.Effort == "" {
			t.Error("Fix missing Effort")
		}
		if fix.Impact == "" {
			t.Error("Fix missing Impact")
		}
	}
}

func TestGenerateFixes_CascadingUpdate(t *testing.T) {
	issue := issues.Issue{
		Type:     issues.IssueCascadingUpdate,
		Severity: issues.SeverityMedium,
	}

	fixes := GenerateFixes(issue)
	if len(fixes) == 0 {
		t.Error("Expected fixes for cascading update issue")
	}
}

func TestGenerateFixes_TimerCascade(t *testing.T) {
	issue := issues.Issue{
		Type:     issues.IssueTimerCascade,
		Severity: issues.SeverityHigh,
	}

	fixes := GenerateFixes(issue)
	if len(fixes) == 0 {
		t.Error("Expected fixes for timer cascade issue")
	}

	// Should include TimelineView suggestion
	hasTimelineView := false
	for _, fix := range fixes {
		if fix.ID == "timeline-view" {
			hasTimelineView = true
		}
	}
	if !hasTimelineView {
		t.Error("Expected TimelineView fix for timer cascade")
	}
}

func TestGenerateRecommendations(t *testing.T) {
	detectedIssues := []issues.Issue{
		{Type: issues.IssueExcessiveRerender, Severity: issues.SeverityHigh},
		{Type: issues.IssueCascadingUpdate, Severity: issues.SeverityMedium},
	}

	recs := GenerateRecommendations(detectedIssues)
	if len(recs) == 0 {
		t.Error("Expected recommendations")
	}

	// Should include @Observable recommendation
	hasObservable := false
	for _, rec := range recs {
		if rec.Title == "Consider using @Observable (iOS 17+)" {
			hasObservable = true
		}
	}
	if !hasObservable {
		t.Error("Expected @Observable recommendation for excessive rerender issues")
	}
}

func TestGetAllFixes(t *testing.T) {
	fixes := GetAllFixes()
	if len(fixes) == 0 {
		t.Error("Expected some fixes from GetAllFixes")
	}

	// Check that all fixes have unique IDs
	seenIDs := make(map[string]bool)
	for _, fix := range fixes {
		if seenIDs[fix.ID] {
			t.Errorf("Duplicate fix ID: %s", fix.ID)
		}
		seenIDs[fix.ID] = true
	}
}

func TestFixHasCodeExamples(t *testing.T) {
	fixes := GetAllFixes()

	hasCodeExamples := false
	for _, fix := range fixes {
		if fix.CodeBefore != "" && fix.CodeAfter != "" {
			hasCodeExamples = true
			break
		}
	}
	if !hasCodeExamples {
		t.Error("Expected at least some fixes to have code examples")
	}
}
