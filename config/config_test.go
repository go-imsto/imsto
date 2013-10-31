package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Logf("confDir: %v", GetConfDir())

	err := LoadConfig("/opt/imsto/config")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("confDir: %v", GetConfDir())
}

func TestGetConfig(t *testing.T) {
	default_db_name := "storage"
	section := ""
	db_name := GetValue(section, "db_name")

	if db_name != default_db_name {
		t.Fatalf("unexpected result from db_name:\n+ %v\n- %v", db_name, default_db_name)
	}
}
