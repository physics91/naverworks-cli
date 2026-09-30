package web

import (
	"encoding/json"
	"testing"
)

func TestTaskUpdatePreservesCategoryAndNullDueDate(t *testing.T) {
	task := Task{ID: "t-51531a20-d2a4-4b5e-bf2b-6a962ecbbdae", CategoryID: "default-category", Title: "before", Content: "keep", AssignorID: "123", Assignees: []Assignee{{ID: "456"}}, StatusOption: "ANY_ONE"}
	title := "after"
	data, err := json.Marshal(taskUpdateBody(task, TaskInput{Title: &title}))
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Title        string
		Content      string
		DueDate      *string
		AssignorID   string
		AssigneeIDs  []string `json:"assigneeIdList"`
		StatusOption string   `json:"taskStatusOption"`
		CategoryID   string
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	if body.DueDate != nil {
		t.Fatalf("an omitted due date became %q instead of retaining null", *body.DueDate)
	}
	if body.Title != title || body.Content != task.Content || body.AssignorID != task.AssignorID || len(body.AssigneeIDs) != 1 || body.AssigneeIDs[0] != "456" || body.StatusOption != task.StatusOption || body.CategoryID != task.CategoryID {
		t.Fatalf("update changed unrelated fields: %s", data)
	}
}

func TestTaskUpdateDueDatePreserveReplaceAndClear(t *testing.T) {
	existing, replacement, empty := "2026-10-01", "2026-10-02", ""
	for _, tc := range []struct {
		name     string
		input    *string
		expected *string
	}{
		{"preserve", nil, &existing},
		{"replace", &replacement, &replacement},
		{"clear", &empty, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(taskUpdateBody(Task{DueDate: &existing}, TaskInput{DueDate: tc.input}))
			if err != nil {
				t.Fatal(err)
			}
			var body struct{ DueDate *string }
			if err := json.Unmarshal(data, &body); err != nil {
				t.Fatal(err)
			}
			if tc.expected == nil {
				if body.DueDate != nil {
					t.Fatal("cleared date was not null")
				}
			} else if body.DueDate == nil || *body.DueDate != *tc.expected {
				t.Fatalf("date not preserved or replaced: %s", data)
			}
		})
	}
}

func TestCreatedTaskNeverAdoptsExistingOrAmbiguousMatches(t *testing.T) {
	title, content := "test", "body"
	body := taskCreateBody("123", TaskInput{Title: &title, Content: &content})
	match := Task{ID: "t-11111111-1111-1111-1111-111111111111", Title: title, Content: content, AssignorID: "123", Assignees: []Assignee{{ID: "123"}}, StatusOption: "ANY_ONE"}
	newTask := match
	newTask.ID = "t-22222222-2222-2222-2222-222222222222"
	otherNew := match
	otherNew.ID = "t-33333333-3333-3333-3333-333333333333"
	oldWithDifferentTitle := match
	oldWithDifferentTitle.Title = "previous title"
	wrongAssignee := newTask
	wrongAssignee.Assignees = []Assignee{{ID: "456"}}
	for _, tc := range []struct {
		name          string
		before, after []Task
		want          string
	}{
		{"existing same title", []Task{match}, []Task{match}, ""},
		{"existing title changed during creation", []Task{oldWithDifferentTitle}, []Task{match, newTask}, newTask.ID},
		{"one newly created match", []Task{match}, []Task{match, newTask}, newTask.ID},
		{"two concurrent identical creates", []Task{match}, []Task{match, newTask, otherNew}, ""},
		{"new task with different assignee", nil, []Task{wrongAssignee}, ""},
		{"no server reflection", nil, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, err := createdTaskID(tc.before, tc.after, body)
			if tc.want == "" {
				if err == nil {
					t.Fatal("accepted an existing or ambiguous creation result")
				}
			} else if err != nil || id != tc.want {
				t.Fatalf("created task mismatch: %q, %v", id, err)
			}
		})
	}
}

func TestCreatedTaskMustMatchRequestedDueDate(t *testing.T) {
	title, content, due := "test", "body", "2026-10-01"
	body := taskCreateBody("123", TaskInput{Title: &title, Content: &content, DueDate: &due})
	task := Task{ID: "t-11111111-1111-1111-1111-111111111111", Title: title, Content: content, AssignorID: "123", Assignees: []Assignee{{ID: "123"}}, StatusOption: "ANY_ONE"}
	if _, err := createdTaskID(nil, []Task{task}, body); err == nil {
		t.Fatal("adopted an unrelated task with a different due date")
	}
	task.DueDate = &due
	if id, err := createdTaskID(nil, []Task{task}, body); err != nil || id != task.ID {
		t.Fatalf("requested due date did not match: %q, %v", id, err)
	}
}
