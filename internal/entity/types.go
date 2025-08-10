package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONB map[string]interface{}

// Implement the `sql.Scanner` interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// Implement the `driver.Valuer` interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type ArrayJSONB []interface{}

// Implement the `sql.Scanner` interface for ArrayJSONB
func (a *ArrayJSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// Implement the `driver.Valuer` interface for ArrayJSONB
func (a ArrayJSONB) Value() (driver.Value, error) {
	return json.Marshal(a)
}
