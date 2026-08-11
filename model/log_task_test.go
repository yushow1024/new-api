package model

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGetUserTaskLogsWithSeparateDatabases(t *testing.T) {
	mainDB, err := gorm.Open(sqlite.Open("file:user-task-logs-main?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open main database: %v", err)
	}
	logDB, err := gorm.Open(sqlite.Open("file:user-task-logs-log?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open log database: %v", err)
	}
	if err := mainDB.AutoMigrate(&Task{}); err != nil {
		t.Fatalf("migrate tasks table: %v", err)
	}
	if err := logDB.AutoMigrate(&Log{}); err != nil {
		t.Fatalf("migrate logs table: %v", err)
	}

	oldDB, oldLogDB, oldLogGroupCol := DB, LOG_DB, logGroupCol
	DB, LOG_DB, logGroupCol = mainDB, logDB, "`group`"
	t.Cleanup(func() {
		DB, LOG_DB, logGroupCol = oldDB, oldLogDB, oldLogGroupCol
	})

	unlinked := &Log{UserId: 7, ModelName: "unlinked", Other: "{}"}
	linked := &Log{UserId: 7, ModelName: "video-model", IsRefunded: true, Other: "{}"}
	otherUserLog := &Log{UserId: 8, ModelName: "other-user", Other: "{}"}
	if err := logDB.Create(unlinked).Error; err != nil {
		t.Fatalf("create unlinked log: %v", err)
	}
	if err := logDB.Create(linked).Error; err != nil {
		t.Fatalf("create linked log: %v", err)
	}
	if err := logDB.Create(otherUserLog).Error; err != nil {
		t.Fatalf("create other user log: %v", err)
	}

	task := &Task{
		UserId:      7,
		TaskID:      "task_linked",
		LogId:       linked.Id,
		ChannelId:   99,
		PrivateData: TaskPrivateData{Key: "secret"},
		ReqData:     json.RawMessage(`{"prompt":"test video"}`),
	}
	otherUserTask := &Task{UserId: 8, TaskID: "task_other", LogId: otherUserLog.Id}
	if err := mainDB.Create(task).Error; err != nil {
		t.Fatalf("create linked task: %v", err)
	}
	if err := mainDB.Create(otherUserTask).Error; err != nil {
		t.Fatalf("create other user task: %v", err)
	}

	items, total, err := GetUserTaskLogs(7, LogTypeUnknown, 0, 0, "", "", 0, 10, "", "", "")
	if err != nil {
		t.Fatalf("get user task logs: %v", err)
	}
	if total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	if len(items) != 2 {
		t.Fatalf("items length = %d, want 2", len(items))
	}
	if items[0].Log.ModelName != "video-model" {
		t.Fatalf("first log model = %q, want video-model", items[0].Log.ModelName)
	}
	if items[0].Log.Id != 1 {
		t.Fatalf("first formatted log id = %d, want 1", items[0].Log.Id)
	}
	if !items[0].Log.IsRefunded {
		t.Fatal("linked log should return is_refunded=true")
	}
	if items[0].Task == nil || items[0].Task.TaskID != "task_linked" {
		t.Fatalf("unexpected linked task: %#v", items[0].Task)
	}
	if items[0].Task.LogId != linked.Id {
		t.Fatalf("task log id = %d, want %d", items[0].Task.LogId, linked.Id)
	}
	if string(items[0].Task.ReqData) != `{"prompt":"test video"}` {
		t.Fatalf("task req_data = %s", items[0].Task.ReqData)
	}
	if items[0].Task.ChannelId != 0 {
		t.Fatalf("user task channel id = %d, want 0", items[0].Task.ChannelId)
	}
	if items[0].Task.PrivateData.Key != "" {
		t.Fatal("private task data should not be loaded")
	}
	if items[1].Log.ModelName != "unlinked" {
		t.Fatalf("second log model = %q, want unlinked", items[1].Log.ModelName)
	}
	if items[1].Log.Id != 2 {
		t.Fatalf("second formatted log id = %d, want 2", items[1].Log.Id)
	}
	if items[1].Task != nil {
		t.Fatalf("unlinked log task = %#v, want nil", items[1].Task)
	}
}
