/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/software-t-rex/monospace/cmd"

	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra/doc"
)

const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
---

`

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		utils.Exit("unable to get the current filename")
	}
	os.Chdir(filepath.Dir(filename))
	// generate markdown documentation
	for _, dir := range []string{"./md", "./site", "./man"} {
		if !utils.FileExistsNoErr(dir) {
			utils.CheckErr(utils.MakeDir(dir))
		}
	}
	fmt.Println("Generating markdown documentation")
	utils.CheckErr(doc.GenMarkdownTree(cmd.RootCmd, "./md"))
	fmt.Println("Generating markdown documentation for website")
	utils.CheckErr(doc.GenMarkdownTreeCustom(cmd.RootCmd, "./site",
		func(filename string) string {
			now := time.Now().Format(time.RFC3339)
			name := filepath.Base(filename)
			base := strings.TrimSuffix(name, path.Ext(name))
			url := "/commands/" + strings.ToLower(base) + "/"
			return fmt.Sprintf(fmTemplate, now, strings.Replace(base, "_", " ", -1), base, url)
		},
		func(s string) string { return s },
	))
	fmt.Println("Generating man pages")
	utils.CheckErr(doc.GenManTree(cmd.RootCmd, &doc.GenManHeader{
		Title:   "monospace",
		Section: "1",
	},
		"./man",
	))
	fmt.Println(utils.Success("Documentation generated successfully"))

}
