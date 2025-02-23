# sercat

`sercat` is a simple utility to interact with a serial port on the command-line.
it integrates with `bash` or equivalent shells to provide meaningful completion
of available ports, speeds and formats.

## Install

There are no prebuilt binaries at this stage. To install, you will need the Go
toolchain from https://go.dev/doc/install. Follow the install instructions, and
make sure that the `bin/` folder for the go modules install location is in your
path.

Then, run the following command:

```sh
go install github.com/maargenton/sercat@latest
```

To activate shell completion, either run the following command in your terminal sessions, or include it in your shell profile:

```sh
eval $(sercat --bash-completion-script)
```

## Usage

```sh
sercat [options] <port> [<baudrate>] [<format>]
```

Use shell completion to easily locate available serial ports and available
baudrate and format options. Use `Ctrl+C` to interrupt the process when done.

Note that the command also transmits any data received on its standard input,
either thought redirection, pipe or directly from the terminal. On the terminal,
characters are accumulated and transmitted only after a newline is received.
