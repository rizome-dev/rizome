package tui

// Copyright (C) 2025 Rizome Labs, Inc.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// ReadMultilineInput reads multi-line input until Ctrl+D
func ReadMultilineInput() (string, error) {
	var lines []string
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// User pressed Ctrl+D, but there might be content on the current line
				// Remove any trailing newline and add the final line if it's not empty
				line = strings.TrimSuffix(line, "\n")
				if len(line) > 0 {
					lines = append(lines, line)
				}
				break
			}
			return "", err
		}

		// Remove the trailing newline for processing
		line = strings.TrimSuffix(line, "\n")
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}
