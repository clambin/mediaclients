package plex

import (
	"context"
)

func (c *Client) GetLibraries(ctx context.Context) ([]Library, error) {
	type response struct {
		Directory []Library `json:"Directory"`
	}
	resp, err := call[response](ctx, c, "/library/sections")
	return resp.Directory, err
}

func (c *Client) GetMovies(ctx context.Context, key string) ([]Movie, error) {
	type response struct {
		Metadata []Movie `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/sections/"+key+"/all")
	return resp.Metadata, err
}

func (c *Client) GetShows(ctx context.Context, key string) ([]Show, error) {
	type response struct {
		Metadata []Show `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/sections/"+key+"/all")
	return resp.Metadata, err
}

func (c *Client) GetSeasons(ctx context.Context, key string) ([]Season, error) {
	type response struct {
		Metadata []Season `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/metadata/"+key+"/children")
	return resp.Metadata, err
}

func (c *Client) GetEpisodes(ctx context.Context, key string) ([]Episode, error) {
	type response struct {
		Metadata []Episode `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/metadata/"+key+"/children")
	return resp.Metadata, err
}

/*
func (c *Client) Raw(ctx context.Context, path string) (any, error) {
	return call[any](ctx, c, path)
}
*/
