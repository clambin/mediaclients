/*
Package plextv provides access to the PlexTV API.

[Config] implements all Plex authentication flows covered in the [Plex API documentation]:
JWT Authentication (Recommended) and Traditional Token Authentication (Legacy). For legacy tokens, both
username/password and PIN flows are supported.

[Config] implements these in an approach similar to [oauth2], though Plex authentication is not compatible with oauth2 itself.

[Client] interacts with plex.tv's API. It uses [Config] to authenticate itself with plex.tv.
Currently, it only supports the /api/v2/user, /api/v2/resources, and /api/v2/devices endpoints.
Plex currently provides no formal documentation for these endpoints, so the implementation may break in the future.
More endpoints may be added in the future.

[Plex API documentation]: https://developer.plex.tv/pms/#section/API-Info/Authenticating-with-Plex
[oauth2]: https://pkg.go.dev/golang.org/x/oauth2
*/
package plextv
