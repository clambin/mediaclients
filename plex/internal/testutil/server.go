package testutil

import (
	"codeberg.org/clambin/go-common/testutils"
	"net/http"
)

func WithToken(token string, next http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Plex-Token") != token {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	}
}

var TestServer = testutils.TestServer{Paths: plexResponses}

var plexResponses = map[string]testutils.Path{
	"/identity": {Body: []byte(`{ "MediaContainer": {
    	"size": 0,
    	"claimed": true,
    	"machineIdentifier": "SomeUUID",
    	"version": "SomeVersion"
  	}}`)},

	"/status/sessions": {Body: []byte(`{ "MediaContainer": {
		"size": 2,
		"Metadata": [
			{ "User": { "title": "foo" },   "Player": { "product": "Plex Web" }, "Session": { "location": "lan"}, "grandparentTitle": "series", "parentTitle": "season 1", "title": "pilot", "type": "episode"},
			{ "User": { "title": "bar" },   "Player": { "product": "Plex Web" }, "Session": { "location": "wan"}, "TranscodeSession": { "throttled": false, "videoDecision": "copy" }, "title": "movie 1" },
			{ "User": { "title": "snafu" }, "Player": { "product": "Plex Web" }, "Session": { "location": "lan"}, "TranscodeSession": { "throttled": true, "speed": 3.1, "videoDecision": "transcode" }, "title": "movie 2" },
			{ "User": { "title": "snafu" }, "Player": { "product": "Plex Web" }, "Session": { "location": "lan"}, "TranscodeSession": { "throttled": true, "speed": 4.1, "videoDecision": "transcode" }, "title": "movie 3" }
		]
	}}`)},

	"/library/sections": {Body: []byte(`{ "MediaContainer": {
		"size": 2,
        "Directory": [
           { "Key": "1", "Type": "movie", "Title": "Movies" },
           { "Key": "2", "Type": "show", "Title": "Shows" }
        ]
    }}`)},

	"/library/sections/1/all": {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "1", "title": "foo" }
        ]
    }}`)},

	"/library/sections/2/all": {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "2", "title": "bar" }
        ]
    }}`)},

	"/library/metadata/200/children": {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "2", "title": "Season 1" }
        ]
    }}`)},

	"/library/metadata/201/children": {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "2", "title": "Episode 1" }
        ]
    }}`)},
}
