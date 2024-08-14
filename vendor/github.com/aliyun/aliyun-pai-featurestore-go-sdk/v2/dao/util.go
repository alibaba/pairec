package dao

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"io"
)

// ColumnValues return dynamic column values by the sql.ColumnType
func ColumnValues(columns []*sql.ColumnType) []interface{} {

	values := make([]interface{}, len(columns))
	for i, column := range columns {
		switch column.DatabaseTypeName() {
		case "TEXT":
			values[i] = &sql.NullString{}
		case "BIGINT", "INT8":
			values[i] = &sql.NullInt64{}
		case "INT", "INT4":
			values[i] = &sql.NullInt32{}
		case "DOUBLE", "FLOAT", "FLOAT8":
			values[i] = &sql.NullFloat64{}
		default:
			values[i] = &sql.NullString{}
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
	default:
		return nil
	}

	return nil
}

func Md5(msg string) string {
	h := md5.New()
	io.WriteString(h, msg)
	return hex.EncodeToString(h.Sum(nil))
}
