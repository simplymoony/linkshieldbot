# linkshieldbot
LinkShieldBot lets you restrict users from joining your group or channel unless they are in another group or channel of
your choice (e.g: User will be able to join Group "A" only if they're member of Group "B").

The code should be easy to extend so feel free to fork and modify it as needed.


## Compatibility
It should be compatible with any OS the Go compiler supports. There are no OS-specific calls in the code.


## Installation

### Compiled binaries
Pre-compiled binaries are available [in the releases page](https://github.com/simplymoony/linkshieldbot/releases).

### From source
First, install the Go compiler from [their website](https://go.dev/dl/). Once you have it installed, you have two options:

- **Option 1 (Recommended)**: Have the `go install` command download, compile and install the latest version for you:

  `go install github.com/simplymoony/linkshieldbot@latest`

  This will install the binary in the directory named by the `GOBIN`
  environment variable, which defaults to `$GOPATH/bin` or `$HOME/go/bin` if the `GOPATH` environment variable is not set.

  TIP: You can use `go env` to view your Go environment.
  
- **Option 2**: Download the source code and compile it manually using the `go build` command.


## Usage
Follow the following steps to get things up and running quickly:
1) Obtain a Telegram Bot API token from [@BotFather](https://t.me/BotFather).
2) Create a `BOT_TOKEN` environment variable with the token you received in step one as value.
3) Run the executable; If this is your first run, the program will output
an error telling you that a config file was generated followed by a path.
4) Open `config.toml` residing at the path you received in step three using a text editor and add
at least one directive below the `[directives]` table. Read the comments to understand the format and
what they do.
5) Save the updated config and re-run the executable. All done!

## Bugs
If you encounter a bug feel free to open an issue or message me on Telegram if you already know me there ;).

Reproducing the error while running in verbose mode (`-verbose` parameter) and providing full logs
would help a lot.