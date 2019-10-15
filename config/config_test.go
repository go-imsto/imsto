package config

import (
	"os"
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
	var tc = func() error {
		return nil
	}
	os.Setenv("IMSTO_CONF", "../apps/demo-config")
	AtLoaded(tc)
	err := Load()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("loaded from: %s", defaultRoot)

	sections := Sections()

	t.Logf("sections: %d", len(sections))

	t.Logf("section default %v", GetSection(defaultSection))

	t.Logf("has section 'demo': %v", HasSection("demo"))

	// t.Fail()
}

func TestGetConfig(t *testing.T) {
	section := defaultSection
	dft_thumb_root := "/opt/imsto/cache/images/"
	thumb_root := GetValue(section, "thumb_root")

	if thumb_root != dft_thumb_root {

		t.Fatalf("unexpected result from thumb_root:\n+ %v\n- %v", thumb_root, dft_thumb_root)
	}

	dft_max_quality := 88
	max_quality := GetInt(section, "max_quality")

	if max_quality != dft_max_quality {

		t.Fatalf("unexpected result from max_quality:\n+ %v\n- %v", max_quality, dft_max_quality)
	}

}
