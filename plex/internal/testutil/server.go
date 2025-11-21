package testutil

import (
	"net/http"

	"codeberg.org/clambin/go-common/testutils"
)

var TestServer = testutils.TestServer{Responses: plexResponses}

var plexResponses = testutils.Responses{
	"/identity": {http.MethodGet: {Body: []byte(`{ "MediaContainer": {
    	"size": 0,
    	"claimed": true,
    	"machineIdentifier": "SomeUUID",
    	"version": "SomeVersion"
  	}}`)}},

	"/status/sessions": {http.MethodGet: {Body: []byte(`{ "MediaContainer": {
		"size": 2,
		"Metadata": [
			{ "User": { "title": "foo" },   "Player": { "product": "Plex Web" }, "Session": { "location": "lan"}, "grandparentTitle": "series", "parentTitle": "season 1", "title": "pilot", "type": "episode"},
			{ "User": { "title": "bar" },   "Player": { "product": "Plex Web" }, "Session": { "location": "wan"}, "TranscodeSession": { "throttled": false, "videoDecision": "copy" }, "title": "movie 1" },
			{ "User": { "title": "snafu" }, "Player": { "product": "Plex Web" }, "Session": { "location": "lan"}, "TranscodeSession": { "throttled": true, "speed": 3.1, "videoDecision": "transcode" }, "title": "movie 2" },
			{ "User": { "title": "snafu" }, "Player": { "product": "Plex Web" }, "Session": { "location": "lan"}, "TranscodeSession": { "throttled": true, "speed": 4.1, "videoDecision": "transcode" }, "title": "movie 3" }
		]
	}}`)}},

	"/library/sections": {http.MethodGet: {Body: []byte(`{ "MediaContainer": {
		"size": 2,
        "Directory": [
           { "Key": "1", "Type": "movie", "Title": "Movies" },
           { "Key": "2", "Type": "show", "Title": "Shows" }
        ]
    }}`)}},

	"/library/sections/1/all": {http.MethodGet: {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "1", "title": "foo" }
        ]
    }}`)}},

	"/library/sections/2/all": {http.MethodGet: {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "2", "title": "bar" }
        ]
    }}`)}},

	"/library/metadata/200/children": {http.MethodGet: {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "2", "title": "Season 1" }
        ]
    }}`)}},

	"/library/metadata/201/children": {http.MethodGet: {Body: []byte(`{ "MediaContainer" : {
        "Metadata": [
           { "guid": "2", "title": "Episode 1" }
        ]
    }}`)}},
}
