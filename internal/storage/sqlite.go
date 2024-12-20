package storage

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements Store interface using SQLite
type SQLiteStore[T any] struct {
	db    *sql.DB
	mu    sync.RWMutex
	table string
}

// TableName interface can be implemented to override table name
type TableName interface {
	TableName() string
}

func NewSQLiteStore[T any](dbPath string) (*SQLiteStore[T], error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("ERROR: Failed to open database: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	var zero T
	table := getTableName(zero)

	store := &SQLiteStore[T]{
		db:    db,
		table: table,
	}

	// Create table if it doesn't exist
	if err := store.initTable(); err != nil {
		db.Close()
		log.Printf("ERROR: Failed to initialize table: %v", err)
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}

	return store, nil
}

func getTableName[T any](data T) string {
	// Check if type implements TableName interface
	if tn, ok := any(data).(TableName); ok {
		return tn.TableName()
	}

	// Otherwise use type name
	t := reflect.TypeOf(data)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return strings.ToLower(t.Name()) + "s"
}

func getFieldsAndTypes[T any]() ([]string, []string, []string, error) {
	var data T
	t := reflect.TypeOf(data)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns, types, fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		sqlTag := field.Tag.Get("sql")
		if sqlTag == "-" {
			continue
		}

		columns = append(columns, strings.ToLower(field.Name))
		fields = append(fields, field.Name)

		// Parse SQL type from tag or infer from Go type
		if sqlTag != "" {
			types = append(types, sqlTag)
		} else {
			sqlType := getSQLType(field.Type)
			types = append(types, sqlType)
		}
	}

	return columns, types, fields, nil
}

func getSQLType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "TEXT"
	case reflect.Int, reflect.Int64:
		return "INTEGER"
	case reflect.Float64:
		return "REAL"
	case reflect.Bool:
		return "BOOLEAN"
	default:
		if t.String() == "time.Time" {
			return "DATETIME"
		}
		return "TEXT"
	}
}

func (s *SQLiteStore[T]) initTable() error {
	columns, types, _, err := getFieldsAndTypes[T]()
	if err != nil {
		return err
	}

	var fields []string
	for i := range columns {
		fields = append(fields, fmt.Sprintf("%s %s", columns[i], types[i]))
	}

	schema := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		%s
	)`, s.table, strings.Join(fields, ",\n\t\t"))

	_, err = s.db.Exec(schema)
	return err
}

func (s *SQLiteStore[T]) Save(data T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	columns, _, fields, err := getFieldsAndTypes[T]()
	if err != nil {
		log.Printf("ERROR: Failed to get fields and types: %v", err)
		return err
	}

	// Create placeholders
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		s.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// Extract values using reflection
	values := make([]interface{}, len(fields))
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i, field := range fields {
		values[i] = v.FieldByName(field).Interface()
	}

	_, err = s.db.Exec(query, values...)
	if err != nil {
		log.Printf("ERROR: Failed to insert data: %v", err)
		return fmt.Errorf("failed to insert data: %w", err)
	}

	return nil
}

// FindBetween returns records between start and end timestamps
func (s *SQLiteStore[T]) FindBetween(start, end interface{}) ([]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := fmt.Sprintf("SELECT * FROM %s WHERE timestamp BETWEEN ? AND ?", s.table)
	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	var results []any
	for rows.Next() {
		var data T
		v := reflect.ValueOf(&data).Elem()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}

		// Skip the ID column
		for i := 1; i < len(columns); i++ {
			field := v.FieldByName(strings.Title(columns[i]))
			if field.IsValid() {
				val := reflect.ValueOf(*(values[i].(*interface{})))
				field.Set(val.Convert(field.Type()))
			}
		}

		results = append(results, data)
	}

	return results, nil
}

func (s *SQLiteStore[T]) Get() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := fmt.Sprintf("SELECT * FROM %s", s.table)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var data T
		v := reflect.ValueOf(&data).Elem()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}

		// Skip the ID column
		for i := 1; i < len(columns); i++ {
			field := v.FieldByName(strings.Title(columns[i]))
			if field.IsValid() {
				val := reflect.ValueOf(*(values[i].(*interface{})))
				field.Set(val.Convert(field.Type()))
			}
		}

		results = append(results, data)
	}

	return results, nil
}

func (s *SQLiteStore[T]) Close() error {
	return s.db.Close()
}
