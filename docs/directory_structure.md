# Directory Structure

This document outlines the directory structure of the `pls7-cli` project.

```
pls7-cli/
├── cmd/
│   └── root.go
├── docs/
│   ├── architecture.md
│   ├── development_plan.md
│   ├── directory_structure.md
│   └── ... (other docs)
├── internal/
│   ├── cli/
│   │   ├── display.go
│   │   ├── format.go
│   │   └── input.go
│   ├── config/
│   │   ├── rules.go
│   │   └── rules_test.go
│   └── util/
│       └── logger.go
├── pkg/
│   ├── poker/
│   │   ├── card.go
│   │   ├── deck.go
│   │   ├── evaluation.go
│   │   ├── odds.go
│   │   ├── rules.go
│   │   └── ... (and test files)
│   └── engine/
│       ├── action.go
│       ├── ai.go
│       ├── betting_limit.go
│       ├── config.go
│       ├── game.go
│       ├── player.go
│       ├── pot.go
│       ├── run.go
│       └── ... (and test files)
├── rules/
│   ├── nlh.yml
│   ├── pls.yml
│   └── pls7.yml
├── main.go
├── go.mod
└── README.md
```

### Role of Each Directory and Package

*   **`main.go`**
    *   The main entry point of the application.
    *   Its only role is to call `cmd.Execute()`.

*   **`cmd/`**
    *   Defines and manages all CLI (Command Line Interface) commands and flags.
    *   `root.go`: Creates the root `pls7` command, defines flags, and contains the main game loop that orchestrates the game flow by calling `pkg/engine` and `internal/cli`.

*   **`rules/`**
    *   Contains YAML files that define the rules for different poker variants. This allows the application to function as a general-purpose poker engine.
    *   `nlh.yml`: Rules for No-Limit Hold'em.
    *   `pls.yml`: Rules for Pot-Limit Sampyeong.
    *   `pls7.yml`: Rules for Pot-Limit Sampyeong 7-or-Better.

*   **`pkg/`**
    *   Contains reusable, domain-specific libraries. Code in this directory is self-contained and has no dependency on the `internal` packages. It can be published and used by other projects.
    *   **`poker/`**: The core poker library. It is a pure library focused on the rules, data models, and evaluation logic of poker.
        *   `rules.go`: Defines the `GameRules` struct, the contract for a poker game's properties.
        *   `card.go`, `deck.go`: Define card and deck structures and operations.
        *   `evaluation.go`: Evaluates hands based on the provided `GameRules`.
        *   `odds.go`: Logic for calculating pot odds, equity, and outs.
    *   **`engine/`**: The game engine. It manages the state and flow of a poker game.
        *   `game.go`: Defines the central `Game` struct, holding the complete state of a running game.
        *   `run.go`: Implements the state machine for a single hand (dealing, processing actions, advancing phases).
        *   `player.go`, `pot.go`, `ai.go`: Define the core components and logic for game progression.
        *   `betting_limit.go`: Implements the strategy for different betting structures (Pot-Limit, No-Limit).

*   **`internal/`**
    *   Contains private application code specific to this CLI project. It is not intended to be imported by other projects.
    *   **`config/`**: Handles loading and parsing rule files from the `/rules` directory into a `poker.GameRules` struct.
    *   **`cli/`**: Manages the "View" and "Input" layers of the CLI.
        *   `display.go`: Renders the `engine.Game` state to the console.
        *   `input.go`: Prompts the user for actions and parses the input.
        *   `format.go`: Provides helper functions for formatting output.
    *   **`util/`**: General-purpose utility functions, like logger initialization.

This structure follows the **Separation of Concerns** principle. The core engine (`pkg/poker` and `pkg/engine`) is completely decoupled from the user interface (`internal/cli`), which would allow for replacing the CLI with a web or GUI front-end while reusing the entire game engine.
