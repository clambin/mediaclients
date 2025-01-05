package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestClientType_getTag(t *testing.T) {
	tests := []struct {
		name string
		clientType
		body any
		want string
	}{
		{
			name:       "sonarr",
			clientType: clientTypeSonarr,
			body: sonarrReleases{
				"v4-stable":  {Version: "1.2.3.4"},
				"v4-nightly": {Version: "1.3.1.1"},
			},
			want: "v1.2.3.4",
		},
		{
			name:       "radarr",
			clientType: clientTypeRadarr,
			body:       []servarrRelease{{Version: "1.2.3.4"}, {Version: "1.2.3.3"}},
			want:       "v1.2.3.4",
		},
		{
			name:       "prowlarr",
			clientType: clientTypeProwlarr,
			body:       []servarrRelease{{Version: "1.2.3.4"}, {Version: "1.2.3.3"}},
			want:       "v1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(tt.body); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			// start test server
			got, err := tt.clientType.getTag(s.URL)
			if err != nil {
				t.Fatalf("getTag() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("getTag() got = %v, want %v", got, tt.want)
			}

			s.Close()
			if _, err = tt.clientType.getTag(s.URL); err == nil {
				t.Errorf("getTag() want error, got nil")
			}
		})
	}
}

func Test_writeFile(t *testing.T) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("failed to create tmp directory: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	err = writeFile(tmp, clientConfig{
		templateVariables: templateVariables{
			Package:    "foo",
			App:        "Foo",
			Tag:        "v1.2.3",
			ApiVersion: "V1",
		},
		clientSource: "client.gen.go",
	})
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tmp, "client.gen.go"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	const want = `package foo

//go:generate oapi-codegen -config config.yaml https://raw.githubusercontent.com/Foo/Foo/refs/tags/v1.2.3/src/Foo.Api.V1/openapi.json
`
	if want != string(got) {
		t.Errorf("writeFile() got = %v, want %v", string(got), want)
	}
}

/*
func TestWriteClientFile(t *testing.T) {
	var buf bytes.Buffer
	vars := templateVariables{
		Package:    "foo",
		App:        "foo",
		Tag:        "v0.1",
		ApiVersion: "V1",
	}
	if err := writeClientFile(&buf, vars); err != nil {
		t.Fatal(err)
	}
	want := `package foo

//go:generate oapi-codegen -config config.yaml https://raw.githubusercontent.com/foo/foo/refs/tags/v0.1/src/foo.Api.V1/openapi.json
`
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}
*/
