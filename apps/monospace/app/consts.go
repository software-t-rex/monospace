package app

import "path/filepath"

const dfltJSPM string = "^pnpm@7.27.0"
const dfltGoModPrfx string = "example.com"
const dfltPreferredOutputMode string = "grouped"

var DfltcfgFilePath string = filepath.Join(".monospace", "monospace.yml")
var DfltHooksDir string = filepath.Join(".monospace", "githooks")
var Version string = "next"
