package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func Test_Main(t *testing.T) {
	tmpdir := t.TempDir()
	configs := []clientConfig{
		{
			templateVariables: templateVariables{Package: "foo", App: "foo", Tag: "v1.2.3"},
			clientSource:      "client1.go",
		},
		{
			templateVariables: templateVariables{Package: "bar", App: "bar", Tag: "v4.5.6"},
			clientSource:      "client2.go",
		},
	}
	var stdout, stderr bytes.Buffer
	Main(&stdout, &stderr, tmpdir, configs)

	if got := stdout.String(); got != "Bump bar to v4.5.6, foo to v1.2.3\n" {
		t.Errorf("stdout = %q; want %q", got, "Bump bar to v4.5.6, foo to v1.2.3\n")
	}
}

func TestClientConfig_currentTag(t *testing.T) {
	cfg := &clientConfig{
		clientSource: filepath.Join("testdata", "client.go"),
	}
	got, err := cfg.currentTag()
	if err != nil {
		t.Fatal(err)
	}
	if got != "v1.2.3" {
		t.Errorf("got %q, want v1.2.3", got)
	}
}

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
			got, err := tt.getTag(s.URL)
			if err != nil {
				t.Fatalf("getTag() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("getTag() got = %v, want %v", got, tt.want)
			}

			s.Close()
			if _, err = tt.getTag(s.URL); err == nil {
				t.Errorf("getTag() want error, got nil")
			}
		})
	}
}

func Test_writeFile(t *testing.T) {
	tmpdir := t.TempDir()
	cfg := clientConfig{
		templateVariables: templateVariables{
			Package:    "foo",
			App:        "Foo",
			Tag:        "v1.2.3",
			ApiVersion: "V1",
		},
		clientSource: "client.gen.go",
	}
	err := writeFile(tmpdir, cfg)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tmpdir, cfg.clientSource))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	gp := filepath.Join("testdata", "client.gen.go.golden")
	if *update {
		if err = os.WriteFile(gp, got, os.ModePerm); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}
	want, err := os.ReadFile(gp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(want, got) {
		t.Errorf("writeFile() got = %v, want %v", string(got), string(want))
	}
}
