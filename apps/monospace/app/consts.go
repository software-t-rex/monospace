package app

import "path/filepath"

const DfltJSPM string = "pnpm@10.11.0"
const DfltGoModPrfx string = "example.com"
const DfltPreferredOutputMode string = "grouped"

var DfltcfgFilePath string = filepath.Join(".monospace", "monospace.yml")
var DfltHooksDir string = filepath.Join(".monospace", "githooks")
var Version string = "next"
