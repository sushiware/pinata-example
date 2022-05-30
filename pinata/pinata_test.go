package pinata_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nasjp/pinata-example/pinata"
)

func TestPinataClientHealthcheck(t *testing.T) {
	client := pinata.New(pinata.PinataAPIKey, pinata.PinataAPISecret)

	if err := client.Healthcheck(); err != nil {
		t.Fatal(err)
	}
}

func TestPinataClientPinJSON(t *testing.T) {
	client := pinata.New(pinata.PinataAPIKey, pinata.PinataAPISecret)

	result, err := client.PinJSON([]byte(`{"hoge":"huga"}`))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestPinataClientPinDir(t *testing.T) {
	client := pinata.New(pinata.PinataAPIKey, pinata.PinataAPISecret)

	metadata := &pinata.PinataMetadata{
		Name: "piyo",
		KeyValues: map[string]string{
			"hoge": "huga",
		},
	}

	dir := "testdata"

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	contents := make([][]byte, 0)
	names := make([]string, 0)

	for _, entry := range files {
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}

		if info.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, info.Name()))
		if err != nil {
			t.Fatal(err)
		}

		contents = append(contents, content)
		names = append(names, info.Name())
	}

	result, err := client.PinDir(contents, names, "images", metadata)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}
