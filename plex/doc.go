/*
Package plex provides a client for interacting with a Plex Media Server.

This package provides several ways to authenticate with a Plex Media Server:

 1. NewPMSClientWithToken creates a new Plex client using a given token. See [Finding an authentication token / X-Plex-Token].
 2. NewPMSClient creates a new Plex client using plex.tv to authenticate itself.

The second option uses the [plextv] package to retrieve a token. It supports both the new (recommended) JWT authentication flow,
and the legacy Credentials and PIN authentication flow.

[Finding an authentication token / X-Plex-Token]: https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/
[plextv]: https://pkg.go.dev/github.com/clambin/mediaclients/plex/plextv
*/
package plex
