/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/spf13/cobra"
)

// checkUpdateCmd represents the checkUpdate command
var checkUpdateCmd = &cobra.Command{
	Use:   "check-update",
	Short: "Will check against monospace github repository for available updates.",
	Long:  `Will check against monospace github repository for available updates.`,
	Run: func(cmd *cobra.Command, args []string) {

		type release struct {
			Name       string `json:"name"`
			Body       string `json:"body"`
			TarballURL string `json:"tarball_url"`
			ZipballURL string `json:"zipball_url"`
			HtmlURL    string `json:"html_url"`
		}

		getLatestRelease := func() (release, error) {
			req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/software-t-rex/monospace/releases/latest", nil)
			if err != nil {
				return release{}, err
			}
			req.Header.Set("Accept", "application/vnd.github.v3+json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return release{}, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return release{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
			var r release
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				return release{}, err
			}
			return r, nil
		}

		var config *app.MonospaceConfig
		if CheckConfigFound(false) {
			config = utils.CheckErrOrReturn(app.ConfigGet())
		}

		latest := utils.CheckErrOrReturn(getLatestRelease())
		if "v"+app.Version == latest.Name {
			fmt.Println(utils.Success("You already have the latest monospace version."))
		} else {
			fmt.Printf("A new monospace version is available: %s\n\n%s\n\nDownload it from: %s\n", latest.Name, latest.Body, latest.HtmlURL)
			binPath, err := os.Executable()
			if err == nil {
				var updateCmd string
				binPath = filepath.ToSlash(binPath)
				if strings.HasSuffix(binPath, "/pnpm/monospace") || strings.HasSuffix(binPath, "/pnpm/monospace.exe") {
					updateCmd = "pnpm add -g @t-rex.software/monospace"
				} else if strings.Contains(binPath, "node_modules/bin/monospace") || strings.Contains(binPath, "npm-packages/bin/monospace") {
					updateCmd = "npm install -g @t-rex.software/monospace"
					if strings.Contains(config.JSPM, "yarn") {
						updateCmd = "yarn global add @t-rex.software/monospace"
					}
				}
				if updateCmd != "" {
					fmt.Printf("\nSeems like you have installed monospace using a package manager.\nYou can update it using this command: %s\n", utils.Bold(updateCmd))
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(checkUpdateCmd)
}
