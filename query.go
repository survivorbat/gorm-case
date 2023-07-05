package gormcase

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const tagName = "gormcase"

func (d *gormCase) queryCallback(db *gorm.DB) {
	// If we only want to use case-insensitivity when explicitly set to true, we back out early if anything's amiss
	settingValue, settingOk := db.Get(tagName)
	if d.conditionalSetting && !settingOk {
		return
	}

	if settingOk {
		if boolValue, _ := settingValue.(bool); !boolValue {
			return
		}
	}

	exp, ok := db.Statement.Clauses["WHERE"].Expression.(clause.Where)
	if !ok {
		return
	}

	for index, cond := range exp.Exprs {
		switch cond := cond.(type) {
		case clause.Eq:
			if d.conditionalTag {
				value := db.Statement.Schema.FieldsByDBName[cond.Column.(string)].Tag.Get(tagName)

				// Ignore if there's no valid tag value
				if value != "true" {
					continue
				}
			}

			value, ok := cond.Value.(string)
			if !ok {
				continue
			}

			condition := fmt.Sprintf("UPPER(%s) = UPPER(?)", cond.Column)

			exp.Exprs[index] = db.Session(&gorm.Session{NewDB: true}).Where(condition, value).Statement.Clauses["WHERE"].Expression
		case clause.IN:
			if d.conditionalTag {
				value := db.Statement.Schema.FieldsByDBName[cond.Column.(string)].Tag.Get(tagName)

				// Ignore if there's no valid tag value
				if value != "true" {
					continue
				}
			}

			var caseCounter int
			var useOr bool

			query := db.Session(&gorm.Session{NewDB: true})

			for _, value := range cond.Values {
				value, ok := value.(string)
				if !ok {
					continue
				}

				caseCounter++

				condition := fmt.Sprintf("UPPER(%s) = UPPER(?)", cond.Column)

				if useOr {
					query = query.Or(condition, value)
					continue
				}

				query = query.Where(condition, value)
				useOr = true
			}

			// Do nothing if no changes were made
			if caseCounter == 0 {
				continue
			}

			// TODO: Determine whether this is efficient
			exp.Exprs[index] = db.Session(&gorm.Session{NewDB: true}).Where(query).Statement.Clauses["WHERE"].Expression
		}
	}
}
