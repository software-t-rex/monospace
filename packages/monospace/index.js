// @ts-nocheck
const fs = require("node:fs")
const path = require("node:path")
const https = require("node:https")
const url = require("node:url")
const architectures = { // only support go first class for now
	arm: "arm",
	arm64: "arm64",
	ia32: "i386",
	x64: "x86_64",
}

const platforms = {
	"darwin": "Darwin",
	"freebsd": "Freebsd",
	"linux": "Linux",
	"openbsd": "Openbsd",
	"win32": "Windows",
}

const supportedDist = [
	"Darwin_arm64",
	"Darwin_x86_64",
	"Freebsd_arm64",
	"Freebsd_armv6",
	"Freebsd_armv7",
	"Freebsd_i386",
	"Freebsd_x86_64",
	"Linux_arm64",
	"Linux_armv6",
	"Linux_armv7",
	"Linux_i386",
	"Linux_x86_64",
	"Openbsd_arm64",
	"Openbsd_armv6",
	"Openbsd_armv7",
	"Openbsd_i386",
	"Openbsd_x86_64",
	"Windows_i386",
	"Windows_x86_64",
]

const get = (options, callback, redirects = 0 ) => {
	let req = https.get({
		...options,
    encoding: null
	}, function (response) {
		if (response.headers.location) {
			if (redirects > 10) {
				return new Error("too many redirects")
			}
			get(url.parse(response.headers.location), callback, redirects+1);
		} else {
			callback(response);
		}
	});
}

const exit = (...e) => {
	console.error(...e)
	process.exit(1)
}
const unarchiveError = (e) => exit("Error: error while unarchiving monospace binary: " + e)
const getPM = () => {
	const ua = process.env.npm_config_user_agent
	if (!ua) {
		return null
	}
	const pm = ua.match(/^(?<name>[^/]+)\/(?<version>[\d.]+)\s+.*$/)?.groups
	return pm || null
}


const pkg = require(path.join(__dirname, "./package.json"))
const arch = architectures[process.arch]
const platform = platforms[process.platform]
const plaformArch = `${platform}_${arch}${arch == "arm" ? process.config.variables?.arm_version : ""}`
if (!supportedDist.includes(plaformArch)) {
	throw new Error(`Unsupported platform or arch ${plaformArch}`)
}
const archiveExt = platform == "Windows" ? "zip" : "tar.gz"
const archive = `monospace_${plaformArch}.${archiveExt}`

// recompose archive path
const archiveUrl = url.parse(`https://github.com/software-t-rex/monospace/releases/download/v${pkg.version}/${archive}`)

// check the script is launched from a package manager
if (getPM() == null) {
	exit(`This script is intended to be run by a node package manager, unable to detect the package manager`)
}
const env = process.env
let installPath = "./bin"
fs.mkdirSync(installPath, { recursive: true })

// if (env && env.npm_config_prefix) {
// 	installPath = path.join(env.npm_config_prefix, 'bin')
// } else {
// 	exit("Can't find binary install path")
// }
console.log("will install binary at ", installPath)
// perform download, unpack and install binary
console.log(`Download binary archive from ${archiveUrl}`)
get(archiveUrl, function (res) {
	if (res.statusCode !== 200) {
		console.log(res.statusCode, res.statusMessage)
		exit("can't download binary: " + res.statusMessage)
	}
	const data = []
	let dataLen = 0
	res.on("data", function (chunk) {
		data.push(chunk)
		dataLen += chunk.length
	})
	console.log("Extracting archive", dataLen)
	switch (archiveExt) {
		case "tar.gz":
			res.on("end", function () {
				require("tar")
					.x({ unlink: true, cwd: installPath }, ["monospace"])
					.on("error", unarchiveError)
					.on("warn", unarchiveError)
					.on("finish", (e) => {
						if (e) {exit(e)}
						console.log("monospace installed to ", installPath)
					})
					.end(Buffer.concat(data), (e) => {
						if (e) {
							exit("Done extracting with error", e)
						} else {
							console.log("Done extracting monospace")
						}
					})
			})
			break;
		case "zip":
			res.on("end", function () {
				const jszip = require("jszip")
				var buf = Buffer.concat(data);
				// here we go !
				jszip.loadAsync(buf)
					.then(function (zip) {
						return zip.file("monospace.exe")?.async("binarystring")
					})
					.then(function (binaryString) {
						if (!binaryString) {
							exit("empty archive")
						} else {
							fs.writeFileSync(path.join(installPath, "monospace.exe"), binaryString, "binary")
						}
					})
					.catch(unarchiveError)
			})
	}

})

