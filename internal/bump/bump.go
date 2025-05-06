package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"
)

func main() {
	for i := range clientConfigs {
		var err error
		if clientConfigs[i].templateVariables.Tag, err = clientConfigs[i].clientType.getTag(""); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to determine tag for %q: %v", clientConfigs[i].clientType, err)
			os.Exit(1)
		}
	}
	Main(os.Stdout, os.Stderr, ".", clientConfigs)
}

func Main(stdout, stderr io.Writer, baseDir string, cfg []clientConfig) {
	changes := make(map[string]string, len(clientConfigs))
	for _, config := range cfg {
		if currentTag, _ := config.currentTag(); currentTag != config.templateVariables.Tag {
			changes[config.templateVariables.App] = config.templateVariables.Tag
			if err := writeFile(baseDir, config); err != nil {
				_, _ = fmt.Fprintf(stderr, "failed to write client file for %q: %v", config.clientType, err)
				os.Exit(1)
			}
		}
	}

	bumps := make([]string, 0, len(changes))
	for app, tag := range changes {
		bumps = append(bumps, app+" to "+tag)
	}
	slices.Sort(bumps)
	if len(bumps) > 0 {
		_, _ = fmt.Fprintln(stdout, "Bump", strings.Join(bumps, ", "))
	}
}

type clientConfig struct {
	templateVariables templateVariables
	clientSource      string
	clientType        clientType
}

var (
	tagRegExp = regexp.MustCompile("/refs/tags/(.*)+/src/")
)

func (c clientConfig) currentTag() (string, error) {
	body, err := os.ReadFile(c.clientSource)
	if err != nil {
		// return nil if there's no clientSource: main will create it
		return "", nil
	}
	if matches := tagRegExp.FindSubmatch(body); len(matches) == 2 {
		return string(matches[1]), nil
	}
	return "", errors.New("failed to parse tag")
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

//go:embed client.go.tmpl
var templates embed.FS

func writeFile(baseDir string, cfg clientConfig) error {
	tmpl, err := templates.ReadFile("client.go.tmpl")
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	f, err := os.OpenFile(filepath.Join(baseDir, cfg.clientSource), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	t, err := template.New("bump").Parse(string(tmpl))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return t.Execute(f, cfg.templateVariables)
}
