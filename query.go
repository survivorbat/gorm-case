package gormcase

import (
	"fmt"
	"strings"

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
		if expression := d.handleExpression(db, cond); expression != nil {
			exp.Exprs[index] = expression
		}
	}
}

func (d *gormCase) handleExpression(db *gorm.DB, cond clause.Expression) clause.Expression {
	switch cond := cond.(type) {
	case clause.Eq:
		if d.conditionalTag {
			columnName, ok := cond.Column.(string)
			if !ok {
				return nil
			}

			table := db.Statement.Schema.Table + "."
			if strings.HasPrefix(columnName, table) {
				columnName = strings.TrimPrefix(columnName, table)
			}

			value := db.Statement.Schema.FieldsByDBName[columnName].Tag.Get(tagName)

			// Ignore if there's no valid tag value
			if value != "true" {
				return nil
			}
		}

		value, ok := cond.Value.(string)
		if !ok {
			return nil
		}

		condition := fmt.Sprintf("UPPER(%s) = UPPER(?)", cond.Column)

		return db.Session(&gorm.Session{NewDB: true}).Where(condition, value).Statement.Clauses["WHERE"].Expression
	case clause.IN:
		if d.conditionalTag {
			columnName, ok := cond.Column.(string)
			if !ok {
				return nil
			}

			table := db.Statement.Schema.Table + "."
			if strings.HasPrefix(columnName, table) {
				columnName = strings.TrimPrefix(columnName, table)
			}

			value := db.Statement.Schema.FieldsByDBName[columnName].Tag.Get(tagName)

			// Ignore if there's no valid tag value
			if value != "true" {
				return nil
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
			return nil
		}

		// TODO: Determine whether this is efficient
		return db.Session(&gorm.Session{NewDB: true}).Where(query).Statement.Clauses["WHERE"].Expression
	case clause.AndConditions:
		for index, nested := range cond.Exprs {
			if expression := d.handleExpression(db, nested); expression != nil {
				cond.Exprs[index] = expression
			}
		}

		return cond
	default:
		return nil
	}
}
