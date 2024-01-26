# Publishing monospace npm package:

* First prepare the monospace release: 
  ```sh
  monospace run doc
  ```
* Check documentation and restore the logo in docs/monospace/cli/md/monospace.md
* Update package json version
  ```sh
  monospace run update-package-json -- --version=v0.0.XX
  ```
* Ensure all repositories are clean and git tag the monospace to the new version
* build the monospace release
  ```sh
  monospace run build
  ```
  it will run ```pnpm run build``` which will:
  * bump the package.json to the latest tag
  * update path to manifest files
  * clean the package directory
  * copy README.md, LICENSE, manifest files to the package folder
  * bundle the installer.js script
  * run npm pack (yes npm not pnpm as we will use npm publish and pnpm/npm behaves slightly differently when packing)

* Create new release on github and publish release files
* Check the package archive content
* run ```npm publish --access public``` this will effectively publish the package to the npm registry
* once done you can run ```pnpm run cleanup``` to remove copied files
