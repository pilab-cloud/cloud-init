package cloudinit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const (
	expectedMetadata = "instance-id: test-instance"
	expectedUserdata = `#cloud-config
users:
  - name: testuser`
)

// setupMockServer sets up a test HTTP server
func setupMockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/meta-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprint(w, expectedMetadata)
	})
	mux.HandleFunc("/user-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprint(w, expectedUserdata)
	})
	mux.HandleFunc("/error-meta-data", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	})
	mux.HandleFunc("/error-user-data", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	})
	return httptest.NewServer(mux)
}

func TestNewNetworkDataSourceConfig(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	cfg, err := NewNetworkDataSourceConfig(server.URL)
	if err != nil {
		t.Fatalf("NewNetworkDataSourceConfig failed: %v", err)
	}
	if cfg.URL != server.URL {
		t.Errorf("expected URL %s, got %s", server.URL, cfg.URL)
	}

	_, err = NewNetworkDataSourceConfig("")
	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}

func TestNetworkDataSourceConfig_GenerateMetadata(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		cfg := &NetworkDataSourceConfig{URL: server.URL}
		metadata, err := cfg.GenerateMetadata()
		if err != nil {
			t.Fatalf("GenerateMetadata failed: %v", err)
		}
		if string(metadata) != expectedMetadata {
			t.Errorf("expected metadata %q, got %q", expectedMetadata, string(metadata))
		}
	})

	t.Run("ServerError", func(t *testing.T) {
		// We need to point to a valid path that our mock server handles, but make the server itself return an error for that path.
		// The mock server has /error-meta-data that returns 500.
		// Correcting the test config to use the base URL and let GenerateMetadata append /meta-data
		// No, GenerateMetadata appends /meta-data. So, the server URL for this specific test config
		// should be such that server.URL + "/meta-data" points to the erroring endpoint.
		// This means the cfg.URL should be server.URL (base) for the erroring handler to be called.
		// The error handler is at /error-meta-data. GenerateMetadata will call server.URL + "/meta-data".
		// This is getting confusing. Let's simplify.
		// The mock server has an endpoint `/error-meta-data` that returns 500.
		// We should test GenerateMetadata by giving it a URL that, when `/meta-data` is appended,
		// results in this error endpoint. Or, more simply, change the config URL to point to the error path directly
		// and have the function try to append /meta-data to *that*. That won't work.

		// The function under test, GenerateMetadata, appends "/meta-data".
		// So, if the server has an endpoint server.URL/error-path/meta-data that returns an error,
		// then cfg.URL should be server.URL/error-path.
		// For this test, we have a handler specifically for server.URL + "/error-meta-data" returning 500.
		// So, cfg.URL should be set to this path *minus* the "/meta-data" part GenerateMetadata adds.
		// This is not ideal. Let's make the mock server return an error for the *actual* /meta-data path.

		// Re-thinking the mock server for errors:
		// A simpler way: have a specific config that points to an endpoint that returns an error.
		// The mock server already has /error-meta-data.
		// GenerateMetadata will call cfg.URL + "/meta-data".
		// So, if cfg.URL is server.URL (which is the base URL of the mock server),
		// and we want to hit the erroring endpoint, we need to adjust the test server.

		// Let's create a dedicated error server for clarity or modify the existing one.
		// For now, let's assume GenerateMetadata is called with a URL that will hit an error.
		// The current mock server has "/error-meta-data" that returns 500.
		// GenerateMetadata appends "/meta-data".
		// To make cfg.GenerateMetadata() hit server.URL/error-meta-data, cfg.URL must be server.URL + "/error-meta-data" - "/meta-data"
		// This is not good.

		// Simpler approach:
		// The GenerateMetadata function *always* appends "/meta-data".
		// So, we need a test server where the path server.URL + "/meta-data" itself returns an error.
		// Let's add a new case to the server for this or use a different server instance.

		// For this test, let's use a config that has a URL that will result in an error.
		// The server has an endpoint `/error-meta-data` which returns a 500.
		// `GenerateMetadata` appends `/meta-data` to its `cfg.URL`.
		// So if `cfg.URL` is `mockServer.URL` and `mockServer` is configured to return error for `/meta-data`, it works.

		// Create a new server that specifically errors on /meta-data
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/meta-data" {
				http.Error(w, "mock server error", http.StatusInternalServerError)
			} else {
				http.NotFound(w, r)
			}
		}))
		defer errorServer.Close()

		errorHandlingCfg := &NetworkDataSourceConfig{URL: errorServer.URL}
		_, err := errorHandlingCfg.GenerateMetadata()
		if err == nil {
			t.Error("expected error for server returning 500, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "status code (500)") {
			t.Errorf("expected error to contain 'status code (500)', got %q", err.Error())
		}
	})

	t.Run("EmptyURL", func(t *testing.T) {
		cfg := &NetworkDataSourceConfig{URL: ""}
		_, err := cfg.GenerateMetadata()
		if err == nil {
			t.Error("expected error for empty URL, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "URL is not configured") {
			t.Errorf("expected error 'URL is not configured', got %q", err.Error())
		}
	})
}

func TestNetworkDataSourceConfig_GenerateUserData(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		cfg := &NetworkDataSourceConfig{URL: server.URL}
		userdata, err := cfg.GenerateUserData()
		if err != nil {
			t.Fatalf("GenerateUserData failed: %v", err)
		}
		if string(userdata) != expectedUserdata {
			t.Errorf("expected userdata %q, got %q", expectedUserdata, string(userdata))
		}
	})

	t.Run("ServerError", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/user-data" {
				http.Error(w, "mock server error", http.StatusInternalServerError)
			} else {
				http.NotFound(w, r)
			}
		}))
		defer errorServer.Close()

		errorHandlingCfg := &NetworkDataSourceConfig{URL: errorServer.URL}
		_, err := errorHandlingCfg.GenerateUserData()
		if err == nil {
			t.Error("expected error for server returning 500, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "status code (500)") {
			t.Errorf("expected error to contain 'status code (500)', got %q", err.Error())
		}
	})

	t.Run("EmptyURL", func(t *testing.T) {
		cfg := &NetworkDataSourceConfig{URL: ""}
		_, err := cfg.GenerateUserData()
		if err == nil {
			t.Error("expected error for empty URL, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "URL is not configured") {
			t.Errorf("expected error 'URL is not configured', got %q", err.Error())
		}
	})
}

func TestNetworkDataSourceConfig_GetVolumeName(t *testing.T) {
	cfg := &NetworkDataSourceConfig{URL: "http://dummy.url"}
	expected := "config-2"
	if name := cfg.GetVolumeName(); name != expected {
		t.Errorf("expected volume name %q, got %q", expected, name)
	}
}

func TestNetworkDataSourceConfig_GetFilePaths(t *testing.T) {
	cfg := &NetworkDataSourceConfig{URL: "http://dummy.url"}
	expected := make(map[string]string)
	if paths := cfg.GetFilePaths(); !reflect.DeepEqual(paths, expected) {
		t.Errorf("expected file paths %v, got %v", expected, paths)
	}
	// Also check if it's non-nil, as make(map[string]string) ensures
	if cfg.GetFilePaths() == nil {
		t.Error("expected non-nil map, got nil")
	}
}
