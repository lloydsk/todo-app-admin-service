package testutil

import (
	"testing"

	"github.com/todo-app/services/admin-service/internal/model/domain"
)

// AssertUserEqual verifies two users are equal (excluding timestamps and version)
func AssertUserEqual(t *testing.T, expected, actual *domain.User, message string) {
	t.Helper()
	
	if expected.ID != actual.ID {
		t.Errorf("%s: ID mismatch - expected %s, got %s", message, expected.ID, actual.ID)
	}
	if expected.Name != actual.Name {
		t.Errorf("%s: Name mismatch - expected %s, got %s", message, expected.Name, actual.Name)
	}
	if expected.Email != actual.Email {
		t.Errorf("%s: Email mismatch - expected %s, got %s", message, expected.Email, actual.Email)
	}
	if expected.Role != actual.Role {
		t.Errorf("%s: Role mismatch - expected %s, got %s", message, expected.Role, actual.Role)
	}
}

// AssertTaskEqual verifies two tasks are equal (excluding timestamps and version)
func AssertTaskEqual(t *testing.T, expected, actual *domain.Task, message string) {
	t.Helper()
	
	if expected.ID != actual.ID {
		t.Errorf("%s: ID mismatch - expected %s, got %s", message, expected.ID, actual.ID)
	}
	if expected.Title != actual.Title {
		t.Errorf("%s: Title mismatch - expected %s, got %s", message, expected.Title, actual.Title)
	}
	if expected.Description != actual.Description {
		t.Errorf("%s: Description mismatch - expected %s, got %s", message, expected.Description, actual.Description)
	}
	if expected.AssigneeID != actual.AssigneeID {
		t.Errorf("%s: AssigneeID mismatch - expected %s, got %s", message, expected.AssigneeID, actual.AssigneeID)
	}
	if expected.Status != actual.Status {
		t.Errorf("%s: Status mismatch - expected %s, got %s", message, expected.Status, actual.Status)
	}
	if expected.Priority != actual.Priority {
		t.Errorf("%s: Priority mismatch - expected %s, got %s", message, expected.Priority, actual.Priority)
	}
}

// AssertCategoryEqual verifies two categories are equal
func AssertCategoryEqual(t *testing.T, expected, actual *domain.Category, message string) {
	t.Helper()
	
	if expected.ID != actual.ID {
		t.Errorf("%s: ID mismatch - expected %s, got %s", message, expected.ID, actual.ID)
	}
	if expected.Name != actual.Name {
		t.Errorf("%s: Name mismatch - expected %s, got %s", message, expected.Name, actual.Name)
	}
	if expected.Description != actual.Description {
		t.Errorf("%s: Description mismatch - expected %s, got %s", message, expected.Description, actual.Description)
	}
	if expected.Color != actual.Color {
		t.Errorf("%s: Color mismatch - expected %s, got %s", message, expected.Color, actual.Color)
	}
}

// AssertTagEqual verifies two tags are equal
func AssertTagEqual(t *testing.T, expected, actual *domain.Tag, message string) {
	t.Helper()
	
	if expected.ID != actual.ID {
		t.Errorf("%s: ID mismatch - expected %s, got %s", message, expected.ID, actual.ID)
	}
	if expected.Name != actual.Name {
		t.Errorf("%s: Name mismatch - expected %s, got %s", message, expected.Name, actual.Name)
	}
	if expected.Color != actual.Color {
		t.Errorf("%s: Color mismatch - expected %s, got %s", message, expected.Color, actual.Color)
	}
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got nil", message)
	}
}


// AssertVersionIncremented verifies that the version was incremented
func AssertVersionIncremented(t *testing.T, originalVersion, newVersion int64, message string) {
	t.Helper()
	
	if newVersion != originalVersion+1 {
		t.Errorf("%s: version not incremented correctly - expected %d, got %d", 
			message, originalVersion+1, newVersion)
	}
}