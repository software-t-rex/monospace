# Publishing monospace npm package:

## First: run ```pnpm run build```
it will:
* bump the package.json to the latest tag
* update path to manifest files
* clean the package directory
* copy README.md, LICENSE, manifest files to the package folder
* bundle the installer.js script
* run npm pack (yes npm not pnpm as we will use npm publish and pnpm/npm behaves slightly differently when packing)

## Verifying the package
then check the archive content

## Then run ```npm publish --access public```
this will effectively publish the package to the npm registry

once done you can run ```pnpm run cleanup``` to remove copied files
