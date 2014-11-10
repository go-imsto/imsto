package storage

import (
	"testing"
)

func TestAppNew(t *testing.T) {
	app := NewApp("demo")
	t.Logf("new app %s: key %s", app.Name, app.ApiKey)
	err := app.Save()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("new app %s: id %d", app.Name, app.Id)

	app1, err := LoadApp(app.ApiKey)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("new app %s: id %d, key %s", app1.Name, app1.Id, app1.ApiKey)
}
