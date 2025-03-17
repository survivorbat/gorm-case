package gormcase

import (
	"gorm.io/gorm"
)

// Compile-time interface check
var _ gorm.Plugin = new(gormCase)

// Option can be given to the New() method to tweak its behaviour
type Option func(gormCase *gormCase)

// SettingOnly makes it so that only queries with the setting 'gormcase' set to true are made case-insensitive.
// This can be configured using db.Set("gormcase", true) on the query.
func SettingOnly() Option {
	return func(gormCase *gormCase) {
		gormCase.conditionalSetting = true
	}
}

// TaggedOnly makes it so that only fields with the tag `gormcase` can be turned into case-insensitive queries,
// useful if you don't want every field to be checked case-insensitively.
func TaggedOnly() Option {
	return func(gormCase *gormCase) {
		gormCase.conditionalTag = true
	}
}

// New creates a new instance of the plugin that can be registered in gorm.
func New(opts ...Option) gorm.Plugin { //nolint:ireturn // accepted for enabling linter
	plugin := &gormCase{}

	for _, opt := range opts {
		opt(plugin)
	}

	return plugin
}

type gormCase struct {
	conditionalTag     bool
	conditionalSetting bool
}

func (d *gormCase) Name() string {
	return "gormcase"
}

func (d *gormCase) Initialize(db *gorm.DB) error {
	return db.Callback().Query().Before("gorm:query").Register("gormcase:query", d.queryCallback)
}
