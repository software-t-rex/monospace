/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/software-t-rex/monospace/cmd"

	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/spf13/cobra/doc"
)

const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
---

`

var md = flag.Bool("md", false, "generate markdown documentation")
var site = flag.Bool("site", false, "generate site templates")
var man = flag.Bool("man", false, "generate manifest documentation")
var help = flag.Bool("help", false, "display help")
var h = flag.Bool("h", false, "display help")

func displayHelp(exitCode int) {
	fmt.Println(`
generate the documentation for monospace commands
doc [flags]

by default generate all the documentation type available, you can pick them individually:
-md, -man, -site

If you don't specify all will be generated. so calling only doc is the same as calling
doc -m -man -site

you can use -h or --help to display this help`)
	os.Exit(exitCode)
}

func main() {
	flag.Parse()
	if *help || *h {
		displayHelp(0)
	}
	if *md == false && *site == false && *man == false {
		*md = true
		*site = true
		*man = true
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		utils.Exit("unable to get the current filename")
	}
	wd := filepath.Dir(filename)
	os.Chdir(wd)
	docDir := filepath.Join(wd, "../../../docs/monospace/cli")
	docDirs := []string{
		filepath.Join(docDir, "md"),
		filepath.Join(docDir, "site"),
		filepath.Join(docDir, "manifest"),
	}
	if *md {
		fmt.Println("Generating markdown documentation")
		if !utils.FileExistsNoErr(docDirs[0]) {
			utils.CheckErr(utils.MakeDir(docDirs[0]))
		}
		utils.CheckErr(doc.GenMarkdownTree(cmd.RootCmd, docDirs[0]))
	}
	if *site {
		fmt.Println("Generating markdown documentation for website")
		if !utils.FileExistsNoErr(docDirs[1]) {
			utils.CheckErr(utils.MakeDir(docDirs[1]))
		}
		utils.CheckErr(doc.GenMarkdownTreeCustom(cmd.RootCmd, docDirs[1],
			func(filename string) string {
				now := time.Now().Format(time.RFC3339)
				name := filepath.Base(filename)
				base := strings.TrimSuffix(name, path.Ext(name))
				url := "/commands/" + strings.ToLower(base) + "/"
				return fmt.Sprintf(fmTemplate, now, strings.Replace(base, "_", " ", -1), base, url)
			},
			func(s string) string { return s },
		))
	}
	if *man {
		fmt.Println("Generating man pages")
		if !utils.FileExistsNoErr(docDirs[2]) {
			utils.CheckErr(utils.MakeDir(docDirs[2]))
		}
		utils.CheckErr(doc.GenManTree(cmd.RootCmd, &doc.GenManHeader{
			Title:   "monospace",
			Section: "1",
		},
			docDirs[2],
		))
	}
	fmt.Println(ui.GetTheme().Success("Documentation generated successfully"))

}
