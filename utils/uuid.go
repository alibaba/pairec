package utils

import uuid "github.com/satori/go.uuid"

func UUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}
