# serve

A lightweight, portable, and dependency-free command-line utility for instantly serving files and directories over HTTP.

This tool is designed to be run as a single binary, making it perfect for quick file sharing across a local network, personal file hosting, and serving payloads during security testing.

![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)
![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)

---

## Features

- **Portable Binary**: Compiles to a single, statically-linked executable with no external dependencies.
- **Direct File & Directory Serving**: Serves content based on the URL path, either streaming a file or listing a directory's contents.
- **Custom Port Configuration**: Run the server on the default port `8080` or specify a custom port as a command-line argument.
- **Graceful Shutdown**: Handles `Ctrl-C` to shut down cleanly, waiting for active transfers to complete.
- **Network Discovery**: Displays accessible local network IP addresses on startup for easy access from other devices.

## Setup

```BASH
curl https://github.com/clubcleaver/serve -o serve
cp ./serve /usr/local/bin/.
export PATH="$PATH:/usr/local/bin"
```

## Usage

The best way to use this tool is as a single, portable binary. Run it from the terminal in any directory you wish to serve, local or remote.
Run the binary from the directory you want to 'serve'.

### Running the Server

- **To run on the default port (`8080`):**

  ```bash
  serve
  ```

- **To run on a custom port (e.g., `8000`):**
  ```bash
  serve 5555
  ```
  The server will start and print the URLs where it is accessible.

### Accessing Content

While it is accessible from a browser, the server is primarily designed for **programmatic access**, allowing scripts to get remote directory information and traverse the file system.

#### **Programmatic Access (via `curl` or Scripts)**

This is the intended use case. You can easily script interactions to list directories and download files.

- **To get a directory listing:**
  The server returns a simple, newline-separated list of files and folders.

  ```bash
  # Request the contents of the 'documents' sub-folder
  curl http://localhost:8080/documents/

  # Output:
  # report.pdf
  # notes.txt
  # images
  ```

- **To download a file:**
  Use the `-O` flag with `curl` to save the file with its original name.

  ```bash
  # Download the report.pdf file from the documents folder
  curl -O http://localhost:8080/documents/report.pdf
  ```

Your scripts can parse the directory listing to discover files and then make subsequent requests to either download those files or traverse into subdirectories.

#### **Human-Readable Access (via Browser)**

You can also use a web browser for simple, manual access.

- **To download a file:** Navigate your browser to its full URL, like `http://localhost:8080/documents/report.pdf`.
- **To see a directory listing:** Navigate to the directory's URL, like `http://localhost:8080/documents/`.

### Stopping the Server

Press **`Ctrl-C`** in the terminal where the server is running. It will perform a graceful shutdown, waiting for active transfers to complete.

## License
This project is licensed under the MIT License.
