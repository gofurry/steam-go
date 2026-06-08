package vdf_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/gofurry/steam-go/addons/vdf"
)

func TestBridgeParseString(t *testing.T) {
	t.Parallel()

	doc, err := vdf.ParseString(`"AppState" { "appid" "730" }`)
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}
	if got := doc.Path("AppState", "appid").AsString(); got != "730" {
		t.Fatalf("appid = %q", got)
	}
}

func TestBridgeMarshalString(t *testing.T) {
	t.Parallel()

	doc := vdf.NewDocument(vdf.NewNode("AppState", vdf.NewValue("appid", "730")))
	got, err := vdf.MarshalString(doc)
	if err != nil {
		t.Fatalf("MarshalString returned error: %v", err)
	}
	if !strings.Contains(got, `"appid"`) || !strings.Contains(got, `"730"`) {
		t.Fatalf("marshaled document does not contain appid: %q", got)
	}
}

func ExampleParseString_appManifest() {
	doc, err := vdf.ParseString(`"AppState" { "appid" "730" "name" "Counter-Strike 2" }`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(doc.Path("AppState", "appid").AsString())
	fmt.Println(doc.Path("AppState", "name").AsString())

	// Output:
	// 730
	// Counter-Strike 2
}

func ExampleParseString_libraryFolders() {
	doc, err := vdf.ParseString(`"libraryfolders" { "0" { "path" "C:\\Steam" } }`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(doc.Path("libraryfolders", "0", "path").AsString())

	// Output:
	// C:\Steam
}
