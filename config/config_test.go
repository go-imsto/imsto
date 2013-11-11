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
	meta_table_suffix := "demo"
	section := ""
	table_suffix := GetValue(section, "meta_table_suffix")

	if table_suffix != meta_table_suffix {

		t.Fatalf("unexpected result from table_suffix:\n+ %v\n- %v", table_suffix, meta_table_suffix)
	}
}
