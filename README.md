# pls7-cli

A simple CLI for Pot Limit Sampyong 7 or Better (PLS7) Poker

## What is PLS7?

Pot Limit Sampyong 7 or Better (PLS7, or Sampyong Hi-Lo) is a variant of poker that combines elements of traditional poker with unique rules and gameplay mechanics. It is played with a standard deck of cards and involves betting, bluffing, and strategic decision-making.

- [Guide - English](https://philipjkim.github.io/posts/20250729-pls7-english-guide/)
- [Guide - Korean](https://philipjkim.github.io/posts/20250724-sampyeong-holdem-guide-v1-4/)

## Installation

This guide will walk you through setting up the Go environment and the project itself.

### 1. Go Language Installation

You need Go version 1.23 or higher to run this application.

#### For macOS Users

The easiest way to install Go on a Mac is by using [Homebrew](https://brew.sh/).

1.  If you don't have Homebrew, open your Terminal and install it with the following command:
    ```bash
    /bin/bash -c "$(curl -fsSL [https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh](https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh))"
    ```
2.  Once Homebrew is installed, install Go with this simple command:
    ```bash
    brew install go
    ```
3.  Verify the installation by checking the version:
    ```bash
    go version
    ```

Alternatively, you can download the official installer from the [Go download page](https://go.dev/dl/).

#### For Windows Users

The recommended way to install Go on Windows is by using the official MSI installer.

1.  Visit the [Go download page](https://go.dev/dl/) and download the MSI installer for Windows.
2.  Run the downloaded installer file. The setup wizard will guide you through the installation process.
3.  The installer will automatically add the Go binary to your system's PATH environment variable.
4.  To verify the installation, open a new Command Prompt or PowerShell window and type:
    ```bash
    go version
    ```

### 2. Project Setup

Once Go is installed on your system, follow these steps to set up the project.

1.  Open your terminal or command prompt.
2.  Clone the repository to your local machine (replace the URL with the actual repository URL):
    ```bash
    git clone [https://github.com/your-username/pls7-cli.git](https://github.com/your-username/pls7-cli.git)
    ```
3.  Navigate into the newly created project directory:
    ```bash
    cd pls7-cli
    ```
4.  Download the necessary dependencies listed in the project:
    ```bash
    go mod tidy
    ```

That's it! You are now ready to run the application.

## Running the App

```bash
# Show help message
go run main.go -h
Starts a new game of Poker (PLS7, PLS, NLH) with 1 player and 5 CPUs.

Usage:
  pls7 [flags]

Flags:
      --dev                 Enable development mode for verbose logging.
  -d, --difficulty string   Set AI difficulty (easy, medium, hard) (default "medium")
  -r, --rule string         Game rule to use (pls7, pls, nlh). (default "pls7")
  -h, --help                help for pls7
      --outs                Shows outs for players if found (temporarily draws fixed good hole cards).

# PLS7, medium AI
go run main.go

# for debugging (not clearing previous output)
go run main.go --dev

# NLH, easy AI, with outs
go run main.go -r nlh -d easy --outs
```

## Creating an Executable

```bash
go build -o pls7 main.go
```

## Testing

```bash
# Simple test
go test ./...

# To run all tests in the project with verbose output
go test -v ./...
```

## 📝 Development Plan

For a detailed step-by-step development plan, please see the [Development Plan](./docs/development_plan.md) document.
