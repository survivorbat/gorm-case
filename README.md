# 🧳 Gorm Case Insensitivity Plugin

[![Go package](https://github.com/survivorbat/gorm-case/actions/workflows/test.yaml/badge.svg)](https://github.com/survivorbat/gorm-case/actions/workflows/test.yaml)

I wanted to have some case-insensitive where-queries and I often use maps for filtering, so I made this plugin to convert
queries into that.

By default, all queries are turned into case-insensitive queries, if you don't want this,
you have 2 options:

- `TaggedOnly()`: Will only change queries on fields that have the `gormcase:"true"` tag
- `SettingOnly()`: Will only change queries on `*gorm.DB` objects that have `.Set("gormcase", true)` set.

If you want a particular query to not be case-insensitive, use `.Set("gormcase", false)`. This works
regardless of configuration.


```go
package main

import (
	gormcase "github.com/survivorbat/gorm-case"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Employee is the example for normal usage
type Employee struct {
	Name string
}

// RestrictedEmployee is the example for gormcase.TaggedOnly()
type RestrictedEmployee struct {
	// Will be case-insensitive in queries
	Name string `gormcase:"true"`

	// Will not be case-insensitive in queries
	Job string
}

func main() {
	// Normal usage
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db.Create(Employee{Name: "jEsSiCa"})
	
	filters := map[string]any{
		"name": "jessica",
	}

	db.Use(gormcase.New())
	db.Model(&Employee{}).Where(filters)

	// Only uses case-insensitive-queries for tagged fields
	db, _ = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db.Create(Employee{Name: "jEsSiCa", Job: "dEvElOpEr"})

	filters := map[string]any{
		"name": "jessica",
		"job": "dEvElOpEr",
	}

	db.Use(gormcase.New(gormcase.TaggedOnly()))
	db.Model(&RestrictedEmployee{}).Where(filters)
}
```

Is automatically turned into a query that looks like this:

```sql
SELECT * FROM employees WHERE UPPER(name) = UPPER('jessica');
```

## ⬇️ Installation

`go get github.com/survivorbat/gorm-case`

## 📋 Usage

```go
package main

import (
    "github.com/survivorbat/gorm-case"
)

func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db.Use(gormcase.New("🍌"))
}

```

## 🔭 Plans

Not much here.
