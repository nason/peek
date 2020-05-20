package peekconfig

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLoadStaticServiceFromFile_EmptyFile(t *testing.T) {
	defer StubConfig(``)()
	service, err := LoadStaticServiceFromFile("apeekdotyaml", "")
	eq(t, err, nil)
	var nilService *SimpleService
	eq(t, service, nilService)
}

func TestLoadStaticServiceFromFile_NoStaticService(t *testing.T) {
	defer StubConfig(`---
version: 2

main:
  type: docker
  port: 80
`)()
	service, err := LoadStaticServiceFromFile("apeekdotyaml", "")
	eq(t, err, nil)
	var nilService *SimpleService
	eq(t, service, nilService)
}

func TestLoadStaticServiceFromFile_MainStaticService(t *testing.T) {
	configTmplStr := `---
version: 2

main:
  type: static
  path: %s
`
	path := "build"

	defer StubConfig(fmt.Sprintf(configTmplStr, path))()
	service, err := LoadStaticServiceFromFile("apeekdotyaml", "")
	eq(t, err, nil)
	if service == nil {
		t.Fatal("Expected a SimpleService returned, got <nil>")
	}
	eq(t, service.Name, "main")
	eq(t, service.Path, path)
}

func TestLoadStaticServiceFromFile_NotMainStaticService(t *testing.T) {
	configTmplStr := `---
version: 2

%s:
  type: static
  path: %s
`
	name := "static-app"
	path := "build"

	config := fmt.Sprintf(configTmplStr, name, path)
	defer StubConfig(config)()
	service, err := LoadStaticServiceFromFile("apeekdotyaml", "")
	eq(t, err, nil)
	if service == nil {
		t.Fatal("Expected a SimpleService returned, got <nil>")
	}
	eq(t, service.Name, name)
	eq(t, service.Path, path)
}

func TestLoadStaticServiceFromFile_MultipleServices(t *testing.T) {
	configTmplStr := `---
version: 2

main:
  type: static
  path: .

%s:
  type: static
  path: %s
`
	name := "static-app"
	path := "build"

	config := fmt.Sprintf(configTmplStr, name, path)
	defer StubConfig(config)()
	service, err := LoadStaticServiceFromFile("apeekdotyaml", name)
	eq(t, err, nil)
	if service == nil {
		t.Fatal("Expected a SimpleService returned, got <nil>")
	}
	eq(t, service.Name, name)
	eq(t, service.Path, path)
}

// Helper functions
func StubConfig(content string) func() {
	orig := ReadConfigFile
	ReadConfigFile = func(fn string) ([]byte, error) {
		return []byte(content), nil
	}
	return func() {
		ReadConfigFile = orig
	}
}

func eq(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}
