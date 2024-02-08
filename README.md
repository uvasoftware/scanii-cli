# Scanii CLI

The Scanii CLI helps you build, test, and manage your [Scanii](https://www.scanii.com) integration right from the terminal.

**With the CLI, you can:**

- Interact with the Scanii API, including analyzing files 
- Start a local mock server to test your API requests

## Installation

The cli is available for macOS, Windows, and Linux. See the [releases page](https://github.com/uvasoftware/scanii-cli/releases)

On MacOS you will need to grant the application permission to run 
```shell
xattr -d com.apple.quarantine /path/to/file
```

A docker container is also provided for running the cli.

```shell
docker run ghcr.io/uvasoftware/scanii-cli:latest
```
Previous container versions can be found [here](https://github.com/uvasoftware/scanii-cli/pkgs/container/scanii-cli). 

## Documentation 

* You should start by configuring the CLI with your API key. You can do this by running `sc configure` and following the prompts.
* Once configured, you can start the test server by running `sc server` and then start sending requests to it.
* All other commands are available by running `sc help`

Here's an example of using the cli to analyze a directory of files:

```shell
./sc files process .                             
⎻⎻⎻⎻
Using endpoint: localhost:4000 and API key: key
success in 4.421542ms
✔ Credentials worked against http://localhost:4000/v2.2/
Processing directory:  .
Found 39 file(s)
Processing files 100% |██████████████████████████████████████████████████████████████████████████| (39/39, 918 it/s)        

✔ Completed in 150.205125ms
✔ Files with findings: 4, unable to process: 0 and successfully processed: 39
Files with findings:
------
  path:           cmd/sc/internal/commands/static/eicar.txt
  id:             fd33128a8da445d3b8308fe6d1588829
  checksum/sah1:  3395856ce81f2b7382dee72602f798b642f14140
  content type:   text/plain; charset=utf-8
  content length: 68 B
  creation date:  2024-02-08T13:38:02.074502Z
  findings:       content.malicious.eicar-test-signature
  metadata:       none
------
------
  path:           internal/engine/testdata/language.txt
  id:             901f13376f584efb8d8ad2bd842751a7
  checksum/sah1:  d487c8d12efa3d46a461a80e4d82e7994e5b5b1b
  content type:   text/plain; charset=utf-8
  content length: 36 B
  creation date:  2024-02-08T13:38:02.080232Z
  findings:       content.en.language.nsfw.0
  metadata:       none
------
------
  path:           cmd/sc/internal/commands/testdata/eicar.txt
  id:             23c379d394694a90b3411a6797885e40
  checksum/sah1:  3395856ce81f2b7382dee72602f798b642f14140
  content type:   text/plain; charset=utf-8
  content length: 68 B
  creation date:  2024-02-08T13:38:02.079376Z
  findings:       content.malicious.eicar-test-signature
  metadata:       none
------
------
  path:           internal/engine/testdata/image.jpg
  id:             3dfbee6ec1864c5496fda69aa1033c45
  checksum/sah1:  7951a43bbfb08fd742224ada280913d1897b89ab
  content type:   image/jpeg
  content length: 631 B
  creation date:  2024-02-08T13:38:02.08033Z
  findings:       content.image.nsfw.nudity
  metadata:       none
------

```

Running the mock server: 
```shell
./sc server         
Scanii test server is starting... 🚀
⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻
• Using API Key: key
• Using API Secret: secret
• Engines with 4 known rules
• Mock server started on http://localhost:4000

Sample usage → curl -u key:secret http://localhost:4000/v2.2/ping
```

#### How the test server works
The test server works by comparing the files sent against a static signature database.The built in signature 
[database](https://github.com/uvasoftware/scanii-cli/blob/main/internal/engine/default.json) includes rules for EICAR and other sample files you can use to test your integration. 

For more sophisticated use cases, you can provide your own configuration to test server with the `--engine` flag. 

```shell

#### Known Limitations
Mock Server
* Callbacks are not supported in the current version of the mock server
* The engine does not really do any analysis, it simply compares files against a signature database
* Requests that fail to retrieve an external location are not saved