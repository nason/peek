package config

import (
	"reflect"
	"testing"
)

func TestParseConfigFile_EmptyFile(t *testing.T) {
	defer StubConfig(``)()
	_, err := ParseConfigFile("somefile")
	eq(t, err.Error(), `unexpected end of JSON input`)
}

func TestParseConfigFile_WrongJSON(t *testing.T) {
	defer StubConfig(`{"blah": 42}`)()
	config, err := ParseConfigFile("somefile")
	eq(t, err, nil)
	eq(t, config, &Config{})
}

func TestParseConfigFile_ValidConfig(t *testing.T) {
	defer StubConfig(validConfig)()
	config, err := ParseConfigFile("somefile")
	eq(t, err, nil)
	if config == nil {
		t.Fatal("expected Config returned. got <nil>")
	}
	auth := config.Auth
	if auth == nil {
		t.Fatal("expected Auth returned in Config. got <nil>")
	}
	eq(t, auth.AccessToken, "fakeaccesstoken")
	eq(t, auth.RefreshToken, "fakerefreshtoken")
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

var validConfig string = `{"auth":{"access_token":"fakeaccesstoken","refresh_token":"fakerefreshtoken","id_token":"","token_type":"Bearer","expires_in":86400}}`
