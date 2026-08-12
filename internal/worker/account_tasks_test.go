package worker

import (
	"testing"

	"zauto/internal/store"
)

func TestTasksForAccountLoginAutoPost(t *testing.T) {
	acc := store.Account{
		ID:                1,
		Name:              "Test",
		AutomationFlow:    "facebook_login_auto_post",
		AutomationEnabled: true,
		AutomationParams:  map[string]interface{}{"post_index": float64(0)},
	}
	tasks, err := TasksForAccount(acc)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || len(tasks[0].Actions) < 2 {
		t.Fatalf("tasks=%+v", tasks)
	}
}
