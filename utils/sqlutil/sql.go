package sqlutil

import (
	"database/sql"
	"reflect"
	"strings"
)

// ColumnValues return dynamic column values by the sql.ColumnType
func ColumnValues(columns []*sql.ColumnType) []interface{} {

	values := make([]interface{}, len(columns))
	for i, column := range columns {
		switch column.ScanType().Kind() {
		case reflect.Int64, reflect.Uint64:
			values[i] = &sql.NullInt64{}
		case reflect.Int32, reflect.Uint32, reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
			values[i] = &sql.NullInt32{}
		case reflect.Bool:
			values[i] = &sql.NullBool{}
		case reflect.String:
			values[i] = &sql.NullString{}
		case reflect.Float32, reflect.Float64:
			values[i] = &sql.NullFloat64{}
		case reflect.Struct:
			if column.ScanType().String() == "time.Time" {
				values[i] = &sql.NullTime{}
			}
		case reflect.Slice:
			if column.ScanType().String() == "sql.RawBytes" {
				values[i] = &sql.NullString{}
			}
		case reflect.Ptr:
			switch column.ScanType().Elem().Kind() {
			case reflect.Int64, reflect.Uint64:
				values[i] = &sql.NullInt64{}
			case reflect.Int32, reflect.Uint32, reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
				values[i] = &sql.NullInt32{}
			case reflect.String:
				values[i] = &sql.NullString{}
			case reflect.Float32, reflect.Float64:
				values[i] = &sql.NullFloat64{}
			case reflect.Struct:
				if column.ScanType().String() == "time.Time" {
					values[i] = &sql.NullTime{}
				}
			}
		default:
			if strings.HasPrefix(column.DatabaseTypeName(), "_") { // array
				values[i] = &[]uint8{}
			} else {
				values[i] = &sql.NullFloat64{}
			}
		}
	}

	return values
}

// ParseColumnValues return true value of column value.
// Retrun nil if the column value is not valid, like nullable.
func ParseColumnValues(value interface{}) interface{} {

	switch v := value.(type) {
	case *sql.NullInt32:
		if v.Valid {
			return v.Int32
		}

	case *sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
	case *sql.NullFloat64:
		if v.Valid {
			return v.Float64
		}
	case *sql.NullBool:
		if v.Valid {
			return v.Bool
		}
	case *sql.NullString:
		if v.Valid {
			return v.String
		}
	case *sql.NullTime:
		if v.Valid {
			return v.Time
		}
	case *[]uint8:
		val := string(*v)
		val = strings.Trim(val, "{}")
		strs := strings.Split(val, ",")
		return strs
	default:
		return nil
	}

	return nil
}

func ColumnValuesByDatabaseTypeName(columns []*sql.ColumnType) []interface{} {

	values := make([]interface{}, len(columns))
	for i, column := range columns {
		switch column.DatabaseTypeName() {
		case "LONG":
			values[i] = &sql.NullInt64{}
		case "INT", "SHORT", "BYTE":
			values[i] = &sql.NullInt32{}
		case "BOOLEAN":
			values[i] = &sql.NullBool{}
		case "FLOAT", "DOUBLE":
			values[i] = &sql.NullFloat64{}
		default:
			values[i] = &sql.NullString{}
		}
	}

	return values
}
