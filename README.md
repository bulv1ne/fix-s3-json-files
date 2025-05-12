# Fix S3 JSON Files

This project processes gzipped JSON files stored in an S3 bucket, fixes formatting issues, and optionally uploads the corrected files back to S3.

The JSON issues are:
- Having multiple json objects in a single line. 
  ```json
  {"key": "value"}{"key": "value"}
  ```
  Becomes:
  ```json
  {"key": "value"}
  {"key": "value"}
  ```
- Replace multiple newlines with a single newline.


## Prerequisites

- [Go 1.24](https://go.dev/dl/) or later installed.
- AWS credentials configured (e.g., via `~/.aws/credentials` or environment variables).
- Access to the S3 bucket containing the files to process.

## Setup

1. Clone the repository:
   ```sh
   git clone git@github.com:bulv1ne/fix-s3-json-files.git
   cd fix-s3-json-files
   ```

2. Update the constants in `main.go`:
    - `bucketName`: Replace with the name of your S3 bucket.
    - `prefixName`: Replace with the prefix path in your bucket (e.g., `raw/domain/live/`).
    - `profileName`: Replace with your AWS profile name.

   Example:
   ```go
   const bucketName = "my-bucket"
   const prefixName = "raw/domain/live/"
   const profileName = "default"
   ```

## Usage

Run the program using the following command:

```sh
go run . --prefix <sub-prefix> [--realrun]
```

### Arguments

- `--prefix`: The sub-prefix to process within the `prefixName`. For example, if `prefixName` is `raw/domain/live/` and you pass `--prefix data/`, the program will process files under `raw/domain/live/data/`.
- `--realrun`: Optional. If provided, the program will upload the fixed files back to S3. Without this flag, it will only print the changes.

### Examples

1. **Dry Run** (only prints changes):
   ```sh
   go run . --prefix data/
   ```

2. **Real Run** (uploads fixed files to S3):
   ```sh
   go run . --prefix data/ --realrun
   ```

## How It Works

1. Lists all objects in the specified S3 bucket and prefix.
2. Downloads and decompresses each gzipped JSON file.
3. Fixes formatting issues in the JSON stream.
4. If changes are detected:
    - In dry run mode, prints the changes.
    - In real run mode, uploads the fixed file back to S3.

## Notes

- Ensure you have the necessary permissions to read and write objects in the S3 bucket.
- Use the `--realrun` flag with caution, as it will overwrite the original files in S3.