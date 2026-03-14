[简体中文](README_CN.md) | **[English](README.md)**

---

![GitHub License](https://img.shields.io/github/license/yhlooo/nfa)
[![GitHub Release](https://img.shields.io/github/v/release/yhlooo/nfa)](https://github.com/yhlooo/nfa/releases/latest)
[![release](https://github.com/yhlooo/nfa/actions/workflows/release.yaml/badge.svg)](https://github.com/yhlooo/nfa/actions/workflows/release.yaml)

# NFA (Not Financial Advice)

A financial trading LLM AI Agent.

> **Note: Any output from this program should not be construed as financial advice.**

## Installation

### Binaries

Download the executable binary from the [Releases](https://github.com/yhlooo/nfa/releases) page, extract it, and place the `nfa` file into any directory in your `$PATH`.

### Docker

Run using the Docker image [`ghcr.io/yhlooo/nfa`](https://github.com/yhlooo/nfa/pkgs/container/nfa):

```bash
docker run -v "${HOME}/.nfa:/root/.nfa" -it --rm ghcr.io/yhlooo/nfa:latest --help
```

### From Sources

Requires Go 1.24.7 or later. Execute the following command to download the source code and build it:

```bash
go install github.com/yhlooo/nfa/cmd/nfa@latest
```

The built binary will be located in `${GOPATH}/bin` by default. Make sure this directory is included in your `$PATH`.

## Usage

Configure models and data sources in `~/.nfa/nfa.json`, see [Configuration Reference](docs/reference/config.md)

Then start an interactive chat session:

```bash
nfa
```

Or run with a single prompt and exit:

```bash
nfa -p "What is P/E ratio?"
```

### Model Selection

Switch models during conversation:

```
/model                             # Interactive model selection
/model deepseek/deepseek-reasoner  # Direct switch

/model :vision  # Switch vision model
```

### Custom Skills

NFA supports extending agent capabilities with custom skills. Create a skill in `~/.nfa/skills/<skill-name>/SKILL.md`:

```markdown
---
name: get-price
description: Get asset price
---

1. Confirm the correct asset code
2. Query the asset price for the last 5 trading days
3. Return price data including date and closing price
```

The agent will automatically load and use these skills when needed.
