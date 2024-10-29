# URL Shortener Service

A lightweight, fast URL shortener service built with Go. This project creates short, obfuscated URLs using XOR-based obfuscation combined with shuffled Base62 encoding, and it provides quick access to URLs by caching frequent lookups with an embedded Least Recently Used (LRU) cache. The URLs and mappings are stored in a SQLite database, and the Go standard library is used for routing and managing endpoints.

> **Note**: The XOR and shuffled Base62 obfuscation approach provides basic obfuscation and is not intended for secure URL shortening. For applications requiring strong security, consider implementing more robust encryption methods, such as AES or Format-Preserving Encryption (FPE).

# Live Demo
A live demo of the URL shortener service is available at: https://pulsarapp.xyz/short_url/

## Features

- **Obfuscation**: Short URL codes are created by applying XOR obfuscation with a configurable secret key, followed by Base62 encoding with a shuffled character set for an added layer of uniqueness and complexity.
- **Efficient Caching**: An embedded LRU cache reduces database hits for frequently accessed URLs, enhancing performance.
- **SQLite Database**: All URL mappings are persistently stored in a lightweight SQLite database.
- **Minimal Dependencies**: Uses Go’s standard library for routing and minimal dependencies for a compact and efficient service.

## Configuration

The application is configured using a JSON file with the following fields:

```json
{
    "xor_secret_key": 12345678,
    "shuffle_key": "your_key",
    "address": "localhost:5000",
    "db_filename": "short_app.db",
    "log_filename": "short_app.log",
    "log_level": "debug",
    "production": false,
    "cache_capacity": 100000
}
```
# Configuration Options
- xor_secret_key: Integer key used for XOR obfuscation of URL IDs.
- shuffle_key: String used to shuffle the Base62 character set, adding obfuscation.
- address: Server address and port for the service (e.g., "localhost:5000").
- db_filename: Filename for the SQLite database file storing URL mappings.
- log_filename: Filename for logging service activity.
- log_level: Sets the logging level ("debug", "info", "warn", "error"). Controls the verbosity of the logs.
- production: Boolean flag (true or false) that determines the logging output:
    - If set to true, logs are written only to the log file specified by log_filename.
    - If set to false, logs are written to both the log file and standard output (stdout), which is helpful during development.
- cache_capacity: The maximun capacity of the cache.

# Building the Project
To build the URL shortener, run the following command inside the cmd directory:
```bash
go build main.go
```
# Using the API
Once the server is running, you can use curl to interact with the API.

## Create a Short URL
To create a new short URL, use a POST request to the /short/post endpoint. Replace the URL in the -d option with the desired URL.
```bash
curl -X POST http://localhost:5000/short/post -d '{"url":"http://yahoo.com/"}'
```
Sample Response:
```json
{"long_url":"http://yahoo.com/","short_url":"ZxD7"}
```
The response will contain the original URL and a newly generated short URL code (in this example, "ZxD7").

## Retrieve the Original URL
You can use either a GET or HEAD request to retrieve the original URL by accessing the /short/get/{short_code} endpoint, replacing {short_code} with the generated code from the POST response.

Example:
```bash
curl --head http://localhost:5000/short/get/ZxD7
```
Sample Response:
```
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: http://yahoo.com/
Date: Fri, 25 Oct 2024 21:07:12 GMT
```
In this response, you’ll receive a 302 Found status with the Location header set to the original URL (http://yahoo.com/ in this example), indicating a redirection to the original URL.

# Running Tests
To run the tests, navigate to the root directory of the project and execute:
```bash
go test -v -race -cover ./...
```
This command performs a thorough test run, checks for data race conditions, and reports code coverage.
