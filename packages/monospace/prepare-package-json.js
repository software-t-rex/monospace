const fs = require('node:fs').promises
const { exec } = require('node:child_process');
const { version } = require('node:os');
const pjson = require("./package.json")

process.chdir(__dirname)

const getManFiles = async () => fs.readdir("../../docs/monospace/cli/manifest").then((dirs) => dirs.map(dir => `./manifest/${dir}`))
const getVersion  = async () => new Promise((resolve, reject) => {
	// check args from command line
	if (process.argv.length > 2) {
		const argversionExp = /^--version=[Vv]?(\d+\.\d+\.\d+)$/
		const v = process.argv.find(arg => !!arg.match(argversionExp))
		if (v) {
			console.log(`set version from command line argument: ${v}`)
			return resolve(v.replace(argversionExp, "$1"))
		}
	}
	console.log(`get version from last git tag, to specify version from command line argument use --version={version}`)
	exec("git describe --tags", (error, stdout, stderr) => {
		if (error) { reject(error) }
		if (!stdout.match(/^v(\d+\.\d+\.\d+)\s*$/)) {
			reject(new Error("Can't read version"))
		}
		resolve(stdout.replace(/^v(\d+\.\d+\.\d+)\s*$/, "$1"))
	})
})

async function main() {
	const version = await getVersion()
	const man = await getManFiles()
	console.log(`Update package.json version from ${pjson.version} to ${version}`)
	console.log(`Update package.json man files to \n- ${man.join("\n- ")}`)
	fs.writeFile("./package.json", JSON.stringify({...pjson, version, man},null, "\t"))
}

main()

