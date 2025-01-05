package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))

	for _, config := range clientConfigs {
		l := logger.With("client", config.clientType.String())
		var err error
		if config.templateVariables.Tag, err = config.clientType.getTag(""); err != nil {
			l.Error("failed to determine tag", "err", err)
			os.Exit(1)
		}
		l.Info("writing client file", "tag", config.templateVariables.Tag)
		if err = writeFile(".", config); err != nil {
			l.Error("failed to write client file", "err", err)
			os.Exit(1)
		}
	}
}

type clientConfig struct {
	templateVariables
	clientSource string
	clientType
}

var clientConfigs = []clientConfig{
	{
		clientType:   clientTypeProwlarr,
		clientSource: "prowlarr/client.go",
		templateVariables: templateVariables{
			Package:    "prowlarr",
			App:        "Prowlarr",
			ApiVersion: "V1",
		},
	},
	{
		clientType:   clientTypeRadarr,
		clientSource: "radarr/client.go",
		templateVariables: templateVariables{
			Package:    "radarr",
			App:        "Radarr",
			ApiVersion: "V3",
		},
	},
	{
		clientType:   clientTypeSonarr,
		clientSource: "sonarr/client.go",
		templateVariables: templateVariables{
			Package:    "sonarr",
			App:        "Sonarr",
			ApiVersion: "V3",
		},
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type clientType int

const (
	_ clientType = iota
	clientTypeSonarr
	clientTypeRadarr
	clientTypeProwlarr
)

func (ct clientType) String() string {
	switch ct {
	case clientTypeSonarr:
		return "sonarr"
	case clientTypeRadarr:
		return "radarr"
	case clientTypeProwlarr:
		return "prowlarr"
	default:
		return "unknown"
	}
}

func (ct clientType) getTag(url string) (string, error) {
	switch ct {
	case clientTypeSonarr:
		return sonarrTag(url)
	case clientTypeRadarr, clientTypeProwlarr:
		return servarrTag(url, ct.String())
	default:
		return "", errors.New("unknown client type")
	}
}

func sonarrTag(url string) (string, error) {
	const sonarrReleasesURL = "https://services.sonarr.tv/v1/releases"
	if url == "" {
		url = sonarrReleasesURL
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var releases sonarrReleases
	if err = json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if release, ok := releases["v4-stable"]; ok {
		return "v" + release.Version, nil
	}
	return "", errors.New("no version found")
}

type sonarrReleases map[string]sonarrRelease

type sonarrRelease struct {
	ReleaseDate    time.Time `json:"releaseDate"`
	ReleaseChannel string    `json:"releaseChannel"`
	Status         string    `json:"status"`
	Version        string    `json:"version"`
	Branch         string    `json:"branch"`
	Changes        struct {
		New   []string `json:"new"`
		Fixed []string `json:"fixed"`
	} `json:"changes"`
	MajorVersion int `json:"majorVersion"`
}

func servarrTag(url string, app string) (string, error) {
	const servarrReleasesURL = "https://%s.servarr.com/v1/update/master/changes"
	if url == "" {
		url = fmt.Sprintf(servarrReleasesURL, app)
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var releases []servarrRelease
	if err = json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if len(releases) == 0 {
		return "", errors.New("no releases found")
	}
	return "v" + releases[0].Version, nil
}

type servarrRelease struct {
	Version     string `json:"version"`
	ReleaseDate string `json:"releaseDate"`
	Filename    string `json:"filename"`
	Url         string `json:"url"`
	Hash        string `json:"hash"`
	Branch      string `json:"branch"`
	Changes     struct {
		New   []string `json:"new"`
		Fixed []string `json:"fixed"`
	} `json:"changes"`
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type templateVariables struct {
	Package    string
	App        string
	Tag        string
	ApiVersion string
}

func writeFile(baseDir string, cfg clientConfig) error {
	f, err := os.OpenFile(filepath.Join(baseDir, cfg.clientSource), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	tmpl, err := template.New("bump").Parse(clientFileTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return tmpl.Execute(f, cfg.templateVariables)
}

const clientFileTemplate = `package {{.Package}}

//go:generate oapi-codegen -config config.yaml https://raw.githubusercontent.com/{{.App}}/{{.App}}/refs/tags/{{.Tag}}/src/{{.App}}.Api.{{.ApiVersion}}/openapi.json
`
