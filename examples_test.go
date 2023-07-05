package gormcase

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ExampleNew() {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	_ = db.Use(New())
	_ = db.Use(New(TaggedOnly()))
	_ = db.Use(New(SettingOnly()))
}
