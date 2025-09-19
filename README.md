[Русский](README.ru.md) | **English**

# fetchtracker

`fetchtracker` is an application for serving files with a download counter, written in Go.

## Main Features

*   **File Serving**: Organizes file "distributions" from folders in a specified directory.
*   **Download Counting**: Tracks statistics for the number of downloads for each file with protection against inflation (using a cookie or an IP + User-Agent pair for 24 hours).
*   **Flexible Templating**: Allows customization of distribution pages using custom `index.html`, Markdown files, and the Go template engine.
*   **Nginx Integration**: Efficiently serves files using the `X-Accel-Redirect` header, which reduces the load on the application.
*   **Redis Storage**: Generated pages and counters are stored in Redis for high performance.
*   **Indexing Control**: The indexing process can be started by sending a `USR1` signal to the process or via a special URL. It uses a blue-green deployment method for seamless updates.
*   **Counter Export**: Ability to export the current counter values to a JSON file by sending a `USR2` signal.
*   **Docker Support**: Easy deployment using Docker Compose.

## How It Works

The application works in conjunction with the Nginx web server. The user creates folders with files in the working directory. These folders become "distributions". The indexing process, which must be run manually when content changes, scans these folders and generates web pages for them, saving them along with counters in Redis. The link to a distribution is a hash of the folder's path.

When a user clicks a "download" link, the application receives a POST request, increments the counter, and sends a special header (e.g., `X-Accel-Redirect`) to Nginx with the real path to the file. Nginx then serves the file directly to the user, minimizing the load on the application.

## Installation and Startup

### Method 1: Quick Start

1.  Make sure you have Docker and `make` installed.
2.  Start the containers with the command:
    ```bash
    make run
    ```
    This command will create a directory with test data in `/tmp/testdata` and launch the containers.
3.  Start the indexing process at the default address http://localhost/index/.
4.  After the index is created, links to the distributions will be displayed.

You can also start the containers with the command:
```bash
SHARE_PATH=/path/to/folders docker compose -f deploy/docker-compose.yml up --build
```
Specify the path to your folders in the `SHARE_PATH` variable.

### Method 2: Manual Installation

1.  Obtain the application binary.
    Download a pre-built binary or build it yourself:
    1.  Ensure you have Go installed.
    2.  Build the application using Make:
        ```bash
        make build
        ```
2.  Install Redis or get the access settings for an existing instance.
3.  Create your folders with distributions.
4.  Create a `config.yml` configuration file with the appropriate settings.
5.  Run the application:
    ```bash
    ./fetchtracker -f config.yml
    ```
6.  Start the indexing process and get the a-ddresses of your distributions.

## Configuration

Application settings are defined in a YAML file (`config.yml`). Here is an example with all available parameters:

```yaml
# Port the application listens on
listen: :10011
# URL for connecting to Redis
redis: redis://localhost/0
log_level: info
indexer:
  # Working directory where distribution folders are located
  work_dir: /data/
  # Number of workers for indexing
  workers: 2
  # Filename for a custom distribution page template
  index_filename: index.html
  # Filename for a description in Markdown format
  desc_filename: description.md
  # Template filename for inserting content from Markdown
  template_filename: template.html
  # Path to the default index.html template (if used)
  index_template: ""
  # Path to the default template.html (if used)
  md_template: ""
  # List of files to be ignored during indexing
  skip_files:
  - description.md
  # File to dump counters to upon receiving the USR2 signal
  dump_filename: /tmp/fetchtracker_counters.json
handler:
  # Base URL used for generating links to distributions
  url: http://127.0.0.1
  # Header for redirecting to Nginx
  header_redirect: X-Accel-Redirect
  # Header from which the user's real IP will be taken
  header_realip: X-Real-IP
```

## Usage

### Creating Distributions

To create a distribution, simply create a new folder in the directory specified by `work_dir` in the configuration file. After running the indexer, a page will be generated for this folder [2].

### Templating

The indexer determines which HTML template to use for generating a distribution page based on the following rules:

1.  If the distribution folder contains an `index.html` file (name is configurable via `index_filename`), it will be used.
2.  If the folder does not contain a `description.md` file (configurable via `desc_filename`), the default template (`index_template` or built-in) is used.
3.  If the folder contains both a `template.html` (configurable via `template_filename`) and a `description.md`, the content from `description.md` is converted from Markdown and inserted into `template.html`.
4.  If the folder only contains `description.md`, it is converted and inserted into the default template (`md_template` or built-in).

### Using Markdown

In `description.md` files, you can use special syntax for links and Frontmatter metadata.

#### Link Syntax

*   `[[test.txt]]`: Generates a link to the file `test.txt`.
*   `[[test.txt|Description]]`: Generates a link to the file with the text "Description".
*   `[[FILES]]`: Inserts a list of links to all files in the distribution.

#### Frontmatter

At the beginning of a Markdown file, you can specify a block with metadata in YAML format to control the display of the distribution.

**Example:**
```markdown
---
title: My Distribution
enabled: true
files: {
    file1.img: Description for the first file
}
---

You can place any description here.

[[FILES]]
```
*   `title`: Replaces the folder name in the page title.
*   `enabled`: `true` or `false`, enables or disables the distribution.
*   `files`: An object where the key is the filename and the value is its description, which will be displayed in the file list.

## Nginx Configuration

For proper operation, Nginx needs to be configured as a reverse proxy.

**Example Configuration:**
```nginx
server {
    listen 80;

    # Distributions
    location /share/ {
        try_files false @backend;
    }

    # For loading counters (statistics in JSON)
    location /stat/ {
        try_files false @backend;
    }

    # File download
    location /file/ {
        try_files false @backend;
    }

    # Indexing. THIS LOCATION MUST BE PROTECTED FROM EXTERNAL ACCESS!
    location /index/ {
        # allow 127.0.0.1;
        # deny all;
        try_files false @backend;
    }

    # Internal location pointing to the real path of the distribution files.
    # The /data/ path should match work_dir from the application config.
    location /data/ {
        internal;
        alias /data/;
    }

    location @backend {
        # Headers to disable caching
	    add_header 'Cache-Control' 'no-cache, no-store, must-revalidate';
    	add_header 'Pragma' 'no-cache';
    	add_header 'Expires' '0';

        proxy_pass http://app:10011; # Address of the Go application
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Development

The project uses a `Makefile` to automate common tasks.

*   `make build`: Builds the application binary `fetchtracker`.
*   `make testdata`: Creates test folders and files in the `/tmp/testdata` directory for debugging and testing.
*   `make build-all`: Builds for all supported platforms.
*   `make release`: Creates release archives.
*   `make clean`: Cleans up build artifacts.
*   `make run`: Runs the application via Docker.
