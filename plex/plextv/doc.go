/*
Package plextv provides two main components.

Firstly, it provides [Config], which implements all Plex authentication flows covered in the [Plex API documentation]:
JWT Authentication (Recommended) and Traditional Token Authentication (Legacy). For legacy tokens, both
username/password and PIN flows are supported.

[Config] implements these in an approach similar to [oauth2], though Plex authentication is not compatible with oauth2 itself.

Secondly, this component provides a [Client] to interact with plex.tv's API. It uses [Config] to authenticate itself with plex.tv.
Currently, it only supports the /api/v2/user and /devices.xml endpoints. More may be added in the future.

[Plex API documentation]: https://developer.plex.tv/pms/#section/API-Info/Authenticating-with-Plex
[oauth2]: https://pkg.go.dev/golang.org/x/oauth2
*/
package plextv
