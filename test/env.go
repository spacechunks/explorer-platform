/*
Explorer Platform, a platform for hosting and discovering Minecraft servers.
Copyright (C) 2024 Yannic Rieger <oss@76k.io>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package test

import (
	"context"
	_ "embed"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/stretchr/testify/require"
	"testing"
)

type Env struct {
	t      *testing.T
	client *hcloud.Client

	servers []*hcloud.Server
}

func New(t *testing.T, hCloudToken string) *Env {
	return &Env{
		t:      t,
		client: hcloud.NewClient(hcloud.WithToken(hCloudToken)),
	}
}

func (e *Env) CreateServer(ctx context.Context) string {
	res, _, err := e.client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name: "e2e_" + RandHexStr(e.t),
		ServerType: &hcloud.ServerType{
			Name: "e2e_" + RandHexStr(e.t),
		},
		Image: &hcloud.Image{
			Name: "cax21",
		},
	})
	require.NoError(e.t, err)
	e.servers = append(e.servers, res.Server)
	return res.Server.PublicNet.IPv4.IP.String()
}

func (e *Env) Cleanup(ctx context.Context) {
	for _, server := range e.servers {
		if _, _, err := e.client.Server.DeleteWithResult(ctx, server); err != nil {
			e.t.Logf("error deleting server: %v", err)
		}
	}
}
