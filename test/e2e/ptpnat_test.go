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

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/spacechunks/platform/nodedev"
	"github.com/spacechunks/platform/test"
	"github.com/stretchr/testify/require"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

func before(t *testing.T, ctx context.Context, testEnv *test.Env) {
	t.Cleanup(func() {
		testEnv.Cleanup(ctx)
	})
	testEnv.Setup(ctx)
	addr := testEnv.CreateServer(ctx)
	test.WaitServerReady(t, addr+":22", 1*time.Minute) // as soon as we can ssh we are ready
	setup(t, testEnv, addr)
}

func TestA(t *testing.T) {
	var (
		testEnv = test.NewEnv(t, "")
		ctx     = context.Background()
	)
	before(t, ctx, testEnv)

}

func TestB(t *testing.T) {
	t.Fail()
}

func setup(t *testing.T, env *test.Env, addr string) {
	err := fs.WalkDir(nodedev.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		data, err := nodedev.Files.ReadFile(path)
		if err != nil {
			return err
		}

		f, err := os.Create("/tmp/" + path)
		if err != nil {
			return err
		}

		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("write: %w", err)
		}

		_, _, err = runCMD(
			fmt.Sprintf("scp -i /tmp/%s -r -o StrictHostKeyChecking=no %s root@%s:/root/%s", env.ID, f.Name(), addr, path),
		)
		if err != nil {
			return fmt.Errorf("exec: %w", err)
		}
		return nil
	})
	require.NoError(t, err)
}

func runCMD(cmd string) (string, int, error) {
	//parts := strings.Split(cmd, " ")
	c := exec.Command("sh", "-c", cmd)
	log.Println(c.String())
	//c.Env = []
	// keep this here in case we want to pipe
	// something into stdin
	//if len(in) > 0 {
	//	c.Stdin = bytes.NewReader(in)
	//}
	// TODO: log files
	var buf bytes.Buffer
	var errbuf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &errbuf

	if err := c.Run(); err != nil {
		log.Println(errbuf.String())
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return "", exit.ExitCode(), fmt.Errorf("non-zero exit: %w", err)
		}
		return "", 0, fmt.Errorf("exit: %w", err)
	}
	return buf.String(), 0, nil
}
