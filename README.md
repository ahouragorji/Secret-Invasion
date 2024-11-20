
## Features

- **Directory Scanning**: Recursively scans the specified directories for sensitive data.
- **File Inclusion/Exclusion**: Configurable file types and specific files to include or ignore.
- **Keyword and Pattern Matching**: Detects secrets based on common keywords or regular expressions.
- **Entropy Analysis**: Identifies potential secrets by analyzing randomness in strings (e.g., API tokens).
- **Environment Variable Fallback**: Can use an environment variable (`secretInvasionConfig`) to specify an alternate configuration file.

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/ahouragorji/Secret-Invasion.git
   cd Secret-Invasion
   ```

2. Build the CLI tool:
   ```bash
   go build -o secret-scanner
   ```

3. Run the tool:
   ```bash
   ./secret-scanner -c config.yaml
   ```

---

## Configuration File

The configuration file (`config.yaml`) controls how the scanner operates. Below are explanations of its key sections:

### Groups
A **group** defines the rules for scanning specific types of sensitive data.

#### Example Configuration

```yaml
groups:
  - name: "AWS Secrets"
    paths:
      include:
        - "./" # Scan the entire current directory
      ignore:
        - "node_modules" # Ignore common dependency folders
        - "logs" # Ignore log folders
        - "test" # Ignore test directories
    files:
      types:
        include:
          - ".env" # Environment variable files
          - ".yaml" # YAML configuration files
        ignore:
          - ".log" # Ignore log files
      names:
        include:
          - "go.mod" # Include Go modules
          - "credentials"
        ignore:
          - "test.json" # Exclude test files
```

### Explanation of Fields

#### Paths
- `include`: Directories to include in the scan (e.g., `./` for the current directory).
- `ignore`: Directories to exclude (e.g., `node_modules`, `logs`).

#### Files
- **Types**: Specify file extensions to scan or ignore (e.g., `.env`, `.yaml`, `.json`).
- **Names**: Include or exclude specific filenames or patterns (e.g., `config.yml`, `credentials`).

#### Patterns
- **Include**: Regular expressions to match filenames containing sensitive keywords like `credentials` or `secret`.
- **Ignore**: Patterns to skip files (e.g., `.*example.*`).

#### Texts
- **Keywords**: List of sensitive keywords to detect, such as `password`, `apiKey`, or `secretAccessKey`.
- **Patterns**: Regular expressions for detecting secret-like values (e.g., JWTs, AWS keys).

#### Entropy
- `enable`: Enable or disable entropy-based detection.
- `threshold`: Strings with entropy above this value are flagged as potential secrets (default: `4.2`).

---

## Usage

Run the scanner with a specific configuration file:

```bash
./secret-scanner -c config.yaml
```

If no configuration file is specified, the tool defaults to:
1. The `secretInvasionConfig` environment variable (if set).
2. The `config.yaml` file in the current directory.

---

## Example Outputs

- **When secrets are found**:
  ```plaintext
  WARNING: Sensitive data detected!
  File: ./config.yaml
  Match: apiKey="AIzaSyD3t6KExampleToken"
  ```

- **When no secrets are found**:
  ```plaintext
  Scan completed. No sensitive data found.
  ```

---

## Contributing

We welcome contributions! Please feel free to:
- Submit bug reports or feature requests.
- Fork the repository and submit pull requests.

---

## License

This project is licensed under the [MIT License](LICENSE). Feel free to use and modify it as needed.

---

## Contact

For any issues or questions, please contact **Ahouragorji** or create an issue in the [GitHub repository](https://github.com/ahouragorji/Secret-Invasion).