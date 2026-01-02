package plex_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlexClient_GetIdentity(t *testing.T) {
	c, s := makeClientAndServer(nil)
	t.Cleanup(s.Close)

	identity, err := c.GetIdentity(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "SomeVersion", identity.Version)
	assert.Equal(t, "pms-client-id-srv1", identity.MachineIdentifier)
}
