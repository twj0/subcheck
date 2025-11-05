# subs-check

`subs-check` is a powerful tool for testing and managing proxy subscription links. It merges multiple subscription sources, tests node health (latency, speed, streaming service availability), and provides the filtered results in various formats.

## Key Features

- **Subscription Management**: Merges multiple subscription links (Clash, V2Ray, Base64, etc.) into a single source.
- **Node Health Checking**:
    - **Availability**: Tests node latency to filter out unresponsive nodes.
    - **Speed Testing**: Measures download speed to identify high-performance nodes.
    - **Streaming & Service Unlocking**: Detects if nodes can access popular services like Netflix, Disney+, YouTube, OpenAI, and more.
- **Node Processing**:
    - **Deduplication**: Removes duplicate nodes based on their properties.
    - **Renaming**: Automatically renames nodes based on their IP geolocation.
- **Subscription Conversion**: Integrates `sub-store` to convert the filtered node list into various formats (Clash, ClashMeta, V2Ray, Sing-Box, etc.).
- **Web Interface**: Provides a dashboard (`/admin`) for monitoring, configuration, and manual checks.
- **Flexible Scheduling**: Supports both interval-based and cron-based scheduling for automatic checks.
- **Persistent Storage**: Saves test results and filtered node lists to various backends, including local filesystem, Cloudflare R2, GitHub Gist, WebDAV, and S3-compatible object storage.
- **Notifications**: Sends status updates and results through over 100 notification channels via Apprise.
- **IP Quality Analysis**: Integrates a shell script to perform in-depth analysis of IP addresses.

## Project Structure

The project consists of two main components:

- **`speed-check/`**: The core Go application that handles subscription processing, testing, and serving the web interface.
- **`ip-quality-check/`**: A standalone shell script for advanced IP address analysis.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (for building from source)

### Configuration

1.  Navigate to the `speed-check` directory:
    ```shell
    cd speed-check
    ```
2.  Copy the example configuration file:
    ```shell
    cp config/config.example.yaml config/config.yaml
    ```
3.  Edit `config/config.yaml` to add your subscription links to the `sub-urls` list and customize other settings as needed.

### Building and Running

You can run the application directly from the source or build a binary.

#### Run from Source

From the `speed-check` directory:
```shell
go run . -f ./config/config.yaml
```

#### Build from Binary

1.  Build the application from the `speed-check` directory:
    ```shell
    make build
    ```
2.  Run the generated executable:
    ```shell
    ./subs-check -f ./config/config.yaml
    ```

## Usage

### Web Interface

Once the application is running, you can access the web interface by navigating to `http://127.0.0.1:8199/admin` in your browser.

### Subscription Links

The application exposes several endpoints to access the filtered and converted subscription lists.

- **Universal Subscription**: `http://127.0.0.1:8299/download/sub`
- **ClashMeta**: `http://127.0.0.1:8299/download/sub?target=ClashMeta`
- **Clash**: `http://127.0.0.1:8299/download/sub?target=Clash`
- **V2Ray**: `http://127.0.0.1:8299/download/sub?target=V2Ray`
- **Sing-Box**: `http://127.0.0.1:8299/download/sub?target=sing-box`

### IP Quality Check

The `ip-quality-check` directory contains a shell script (`ip.sh`) for more detailed IP analysis. You can run it with an IP address as an argument.

```shell
cd ip-quality-check
./ip.sh <IP_ADDRESS>
```

## Acknowledgements

This project is built upon the work of several open-source projects, including [cmliu](https://github.com/cmliu), [Sub-Store](https://github.com/sub-store-org/Sub-Store), [bestruirui](https://github.com/bestruirui/BestSub), and [iplark](https://iplark.com/).

## Disclaimer

This tool is for learning and research purposes only. Users should bear all risks and comply with relevant laws and regulations.
