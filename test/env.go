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
	"encoding/pem"
	"os"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
)

type Env struct {
	ID string

	t      *testing.T
	client *hcloud.Client

	servers  []*hcloud.Server
	sshKeyID int64
}

func NewEnv(t *testing.T, hcloudToken string) *Env {
	return &Env{
		ID:     RandHexStr(t),
		t:      t,
		client: hcloud.NewClient(hcloud.WithToken(hcloudToken)),
	}
}

// Setup the environment. Currently, this is:
// * generating and configuring an ed25519 key pair
func (e *Env) Setup(ctx context.Context) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(e.t, err)

	sshpub, err := ssh.NewPublicKey(pub)
	require.NoError(e.t, err)

	sshpriv, err := ssh.MarshalPrivateKey(priv, "")
	require.NoError(e.t, err)

	hcloudKey, _, err := e.client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      "e2e-" + e.ID,
		PublicKey: string(ssh.MarshalAuthorizedKey(sshpub)),
	})
	require.NoError(e.t, err)

	require.NoError(e.t, os.WriteFile(e.PrivateKeyPath(), pem.EncodeToMemory(sshpriv), 0600))

	// hcloud api requires us to use the key id instead of its name when creating a server,
	// so in order to not having to do an extra api call simply set this value here.
	e.sshKeyID = hcloudKey.ID
}

func (e *Env) CreateServer(ctx context.Context) string {
	if e.sshKeyID == 0 {
		e.t.Fatal("no ssh key configured")
	}

	name := "e2e-" + RandHexStr(e.t)
	res, _, err := e.client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name: name,
		SSHKeys: []*hcloud.SSHKey{
			{
				ID: e.sshKeyID,
			},
		},
		ServerType: &hcloud.ServerType{
			Name: "cax21",
		},
		Image: &hcloud.Image{
			Name: "debian-12",
		},
	})
	require.NoError(e.t, err)

	e.servers = append(e.servers, res.Server)
	return res.Server.PublicNet.IPv4.IP.String()
}

func (e *Env) PrivateKeyPath() string {
	return "/tmp/" + e.ID
}

func (e *Env) SSHUser() string {
	return "root"
}

func (e *Env) Cleanup(ctx context.Context) {
	if _, err := e.client.SSHKey.Delete(ctx, &hcloud.SSHKey{
		ID: e.sshKeyID,
	}); err != nil {
		e.t.Logf("error deleting ssh key %s: %v", e.ID, err)
	}

	for _, server := range e.servers {
		if _, _, err := e.client.Server.DeleteWithResult(ctx, server); err != nil {
			e.t.Logf("error deleting server: %v", err)
		}
	}
}
