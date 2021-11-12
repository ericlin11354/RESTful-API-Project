package database

import (
	"testing"
)

func TestInitDb(t *testing.T) {
	InitDb("production", "testing")
	InitDb("development", "testing")
	InitDb("local", "testing")
}
