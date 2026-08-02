package model

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestLogRequestResponseDataMigrationAndPersistence(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:log-data-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	if err := db.AutoMigrate(&Log{}); err != nil {
		t.Fatalf("migrate logs table: %v", err)
	}
	if !db.Migrator().HasColumn(&Log{}, "req_data") {
		t.Fatal("logs.req_data was not migrated")
	}
	if !db.Migrator().HasColumn(&Log{}, "res_data") {
		t.Fatal("logs.res_data was not migrated")
	}

	requestData := LogData(`{"model":"gpt-image-1","prompt":"test"}`)
	responseData := LogData(`{"data":[{"b64_json":"` + strings.Repeat("a", 128*1024) + `"}]}`)
	created := &Log{ReqData: requestData, ResData: responseData}
	if err := db.Create(created).Error; err != nil {
		t.Fatalf("create log: %v", err)
	}

	var stored Log
	if err := db.First(&stored, created.Id).Error; err != nil {
		t.Fatalf("read log: %v", err)
	}
	if stored.ReqData != requestData {
		t.Fatalf("req_data mismatch: got %q", stored.ReqData)
	}
	if stored.ResData != responseData {
		t.Fatalf("res_data length = %d, want %d", len(stored.ResData), len(responseData))
	}
}
