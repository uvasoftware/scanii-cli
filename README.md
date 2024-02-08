# Scanii CLI

The Scanii CLI helps you build, test, and manage your [Scanii](https://wwww.scanii.com) integration right from the terminal.

**With the CLI, you can:**

- Interact with the Scanii API, including analyzing files 
- Start a local mock server to test your API requests

## Installation

The cli is available for macOS, Windows, and Linux. See the [releases page](https://github.com/uvasoftware/scanii-cli/releases)

On MacOS you will need to grant the application permission to run 
```shell
xattr -d com.apple.quarantine /path/to/file
```

## Documentation 

* You should start by configuring the CLI with your API key. You can do this by running `sc configure` and following the prompts.
* Once configured, you can start the mock server by running `sc server` and then start sending requests to it.
* All other commands are available by running `sc help`

### Known Limitations
#### Mock Server
* Callbacks are not supported in the current version of the library.
* The engine does not really do any analysis, it simply compares files against a signature database
* Requests that fail to retrieve an external location are not saved