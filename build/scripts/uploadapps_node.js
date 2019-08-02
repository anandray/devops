#!/usr/bin/env node

const fs = require('fs');
const UtilityName = "wcloud"
const wolkjsdir = "/Users/rodney/src/github.com/wolkdb//wolkjs/";
const appdir = wolkjsdir + "apps";
const WOLK = require(wolkjsdir + "/index.js").FS;
const cloudstore = new WOLK();

var setupApp = async function( owner ) {
	return await cloudstore.createAccount(owner)
		.then( () => {
			return;
			//Creating Already Sets it to Default Account -- Change?
			console.log("attempting to create: " + owner + "/js")
			return cloudstore.createCollection(owner, "js")
		})
		.then( () => {
			return
			let wolkjsLocalFile = appdir + "/wolk.js";
			let wolkjsDestination = "wolk://" + owner + "/js/wolk.js";
			if (!fs.existsSync(wolkjsLocalFile)) {
				console.log("no such file:", wolkjsLocalFile)
				throw new Error("The file attempting to be uploaded doesn't exist: " + wolkjsLocalFile)
			}
			var wolkjsLocalFileContent = fs.readFileSync(wolkjsLocalFile);
			return cloudstore.setKey(owner, "js", "wolk.js", wolkjsLocalFileContent)
			.then( (txhash) => {
				console.log("{\"txhash\":\"" + txhash + "\"}\n")
			})
			.catch( (err) => {
				console.log("ERROR: PutFile - " + err + " owner : " + owner + " coll: js | key: wolk.js" )
			})
		})
		.catch( err => {
			console.log("ERROR: setupApp - " + err)
		})
}

async function main() {
	var owner = "rodneyapp";
	var ports = [443, 81, 82, 83, 84, 85];
	ports = [81];
	var apps = new Map();
	apps.set("explorer", ["home.html", "navbar.html", "txns.html", "block.html", "footer.html"])
	for (var portIndex = 0; portIndex < ports.length; portIndex ++) {
		var port = ports[portIndex]
		var server = "c0.wolk.com"
		var currentScheme = ""
		if ( port == 81 ) {
			currentScheme = "http://"
		} else {
			currentScheme = "https://";
		}
		var currentProvider = currentScheme + server + ":" + port.toString(10)
		cloudstore.setProvider( currentProvider )
		console.log("curProv: " + cloudstore.provider)
		await setupApp( owner )
		.then( () => {
			return
			for ( var appIndex=0; appIndex < apps.length; appIndex++) {
				var app = apps[appIndex]
				.then( () => {
					return cloudstore.createCollection(owner, app)
				})
				.then( () => {
					var fileList = apps.get("explorer")
					for ( var fileIndex=0; fileIndex < fileList.length; fileIndex++) {
						var currentFile = fileList[fileIndex];
						var fileLocation = appdir + "/" + currentFile;
						let destination = "wolk://" + owner + "/" + app + "/" + currentFile;
						if (!fs.existsSync(fileLocation)) {
							console.log("no such file:", fileLocation)
							throw new Error("The file attempting to be uploaded doesn't exist: " + fileLocation)
						}
						var fileContent = fs.readFileSync(fileLocation);
						cloudstore.setKey(owner, app, currentFile, fileContent)
					}
				})
			}
		})
	}
}

main();

/*
#!/usr/bin/php
<?php

$appdir = "/root/go/src/github.com/wolkdb/wolkjs/apps";

function myexec($cmd, $run) {
	echo "$cmd\n";
	if ( $run ) {
		$output = array();
		exec($cmd, $output);
		if ( count($output) > 0 ) {
			print_r($output);
		}
	}
}

function exec_cmds($cmds, $run = false) {
	foreach ($cmds as $cmd) {
    echo "$cmd\n";
	}
}

function uploadapps($owner = "app", $app = "explorer", $files = null)
{
  global $appdir, $port;
  $hp = "-httpport=$port";
  $cmds = array();
  $cmds[] = "wcloud $hp --waitfortx  mkdir wolk://$owner/$app";
  foreach ($files as $i => $fn) {
    $cmds[] = "wcloud $hp  put $appdir/$app/$fn   wolk://$owner/$app/$fn";
  }
  return $cmds;
}

$owner = "app";
$ports = array(443, 81, 82, 83, 84, 85);
$apps = array("explorer" => array("home.html", "navbar.html", "txns.html", "block.html", "footer.html"));
foreach ($ports as $port) {
  $cmds = setup($owner);
  exec_cmds($cmds);
  foreach ($apps as $app => $files) {
    $cmds = uploadapps($owner, $app, $files);
    exec_cmds($cmds);
  }
}
?>
*/
