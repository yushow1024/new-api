package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// LogData stores potentially large request/response JSON payloads. MySQL TEXT
// is limited to 64 KiB, so use LONGTEXT there while keeping the portable TEXT
// type for PostgreSQL and SQLite.
type LogData string

func (LogData) GormDataType() string {
	return "text"
}

func (LogData) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	if db.Dialector.Name() == "mysql" {
		return "LONGTEXT"
	}
	return "TEXT"
}
