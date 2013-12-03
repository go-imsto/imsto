package config

import (
	"testing"
)

const (
	defaultRoot    = "/etc/imsto"
	defaultSection = ""
)

func TestRoot(t *testing.T) {
	SetRoot(defaultRoot)
	t.Logf("Root: %v", Root())
}

func TestLoadConfig(t *testing.T) {
	err := Load()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("loaded from: %s", defaultRoot)

	sections := Sections()

	t.Logf("sections: %d", len(sections))

	t.Logf("section default", GetSection(defaultSection))

	t.Logf("has section 'demo': %s", HasSection("demo"))
}

func TestGetConfig(t *testing.T) {
	section := defaultSection
	dft_thumb_path := "/thumb"
	thumb_path := GetValue(section, "thumb_path")

	if thumb_path != dft_thumb_path {

		t.Fatalf("unexpected result from thumb_path:\n+ %v\n- %v", thumb_path, dft_thumb_path)
	}

	dft_max_quality := 88
	max_quality := GetInt(section, "max_quality")

	if max_quality != dft_max_quality {

		t.Fatalf("unexpected result from max_quality:\n+ %v\n- %v", max_quality, dft_max_quality)
	}

}
