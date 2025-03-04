# Issue Scouter

Issue Scouter is a GitHub Action that helps you find issues you can contribute to within specified repositories.

## Usage

### 1. Create a Repository

To use Issue Scouter, you need to create a dedicated GitHub repository where the generated issue list will be stored.

cf. https://github.com/ymtdzzz/my-issue-scouter

### 2. Create a Configuration File

Create a configuration file (e.g., `config.yml`) to specify the repositories and labels you want to track.

```yaml
repositories:
  OpenTelemetry:
    - https://github.com/open-telemetry/opentelemetry-ruby
    - https://github.com/open-telemetry/opentelemetry-ruby-contrib
  Gem:
    - https://github.com/lostisland/faraday
    - https://github.com/fog/fog-google
labels:
 - "help wanted"
 - "good first issue"
```

### 2. Set Up the Workflow

Create a GitHub Actions workflow (e.g., `.github/workflows/issue-scouter.yml`) to run Issue Scouter periodically.

```yaml
name: Run Issue Scouter

on:
  workflow_dispatch:
  schedule:
    - cron: '0 15 * * *' # 00:00:00 JST

jobs:
  run-issue-scouter:
    runs-on: ubuntu-latest
    permissions:
      contents: write # issue-scouter needs write permission to commit changes in your repository
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
      - name: Run Issue Scouter
        uses: ymtdzzz/issue-scouter@v0.0.5 # use the latest version of issue-scouter
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          dry_run: "false"
          config_file: "config.yml"
```

### 3. View the Generated Issue List

After execution, an issue list will be generated in your repository. You can check an example output at https://github.com/ymtdzzz/my-issue-scouter .

## Contributing

Contributions are welcome! Feel free to submit issues and pull requests to improve Issue Scouter.
