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

package image

import (
	"fmt"
	"log"

	ociv1 "github.com/google/go-containerregistry/pkg/v1"
)

func Repack(src ociv1.Image, rootPath string, path string) (ociv1.Image, error) {
	// for each variant create a new image
	// using the pushed one as its base
	log.Printf("extract flavor=%s\n", path)
	files, err := UnpackDir(src, path)
	log.Printf("extracted flavor=%s\n", path)
	if err != nil {
		return nil, fmt.Errorf("extract dir: %w", err)
	}
	for i := range files {
		// adjust absolute path so flavor files end up
		// in the server root directory when appending the layer.
		files[i].AbsPath = fmt.Sprintf("%s%s", rootPath, files[i].StrippedPath)
	}
	log.Printf("append flavor=%s\n", path)
	img, err := AppendLayerFromFiles(src, files)
	log.Printf("appended flavor=%s\n", path)
	if err != nil {
		return nil, fmt.Errorf("append layer: %w", err)
	}
	return img, nil
}
