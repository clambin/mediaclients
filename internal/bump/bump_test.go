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
	cfg := clientConfig{
		templateVariables: templateVariables{
			Package:    "foo",
			App:        "Foo",
			Tag:        "v1.2.3",
			ApiVersion: "V1",
		},
		clientSource: "client.gen.go",
	}
	err = writeFile(tmp, cfg)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tmp, cfg.clientSource))
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
