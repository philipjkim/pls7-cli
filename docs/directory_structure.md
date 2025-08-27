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
│   ├── game/
│   │   ├── action.go
│   │   ├── ai.go
│   │   ├── betting_limit.go
│   │   ├── config.go
│   │   ├── game.go
│   │   ├── player.go
│   │   ├── pot.go
│   │   ├── run.go
│   │   └── ... (and test files)
│   └── util/
│       └── logger.go
├── pkg/
│   └── poker/
│       ├── card.go
│       ├── deck.go
│       ├── evaluation.go
│       ├── odds.go
│       ├── rules.go
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
    *   `root.go`: Creates the root `pls7` command, defines all flags (e.g., `--rule`, `--difficulty`), and contains the main game loop that orchestrates the entire game flow.

*   **`rules/`**
    *   Contains YAML files that define the rules for different poker variants. This allows the application to function as a general-purpose poker engine.
    *   `nlh.yml`: Rules for No-Limit Hold'em.
    *   `pls.yml`: Rules for Pot-Limit Sampyeong.
    *   `pls7.yml`: Rules for Pot-Limit Sampyeong 7-or-Better.

*   **`pkg/`**
    *   Contains reusable, domain-specific libraries. Code in this directory is self-contained and has no dependency on the `internal` packages. It could theoretically be published and used by other projects.
    *   **`poker/`**: The core poker engine. It is a pure library with no knowledge of the CLI or the specific game flow.
        *   `rules.go`: Defines the `GameRules` struct, which is populated from the YAML files. This is the contract that defines a poker game's properties.
        *   `card.go`, `deck.go`: Define card and deck structures and basic operations.
        *   `evaluation.go`: The most complex part of the engine. It evaluates hands based on the provided `GameRules` (e.g., standard hands, skip straights, low hands).
        *   `odds.go`: Logic for calculating pot odds, equity, and outs.

*   **`internal/`**
    *   Contains all the private application code that is specific to this project. It is not intended to be imported by other projects.
    *   **`config/`**: Handles loading and parsing the rule files from the `/rules` directory.
        *   `rules.go`: Contains the logic to read a YAML file and unmarshal it into a `poker.GameRules` struct.
    *   **`cli/`**: Manages the "View" and "Input" layers of the CLI.
        *   `display.go`: Renders the game state (board, players, pot) to the console.
        *   `input.go`: Prompts the user for actions (check, bet, raise, fold) and parses the input.
        *   `format.go`: Provides helper functions for formatting output, like adding commas to numbers.
    *   **`game/`**: The application's orchestrator. It connects the `poker` engine with the `cli` and manages the game state and turn-based flow.
        *   `game.go`: Defines the central `Game` struct, which holds the complete state of a running game (players, deck, phase, pot, etc.).
        *   `run.go`: Implements the state machine for a single hand (dealing, processing actions, advancing phases).
        *   `player.go`: Defines the `Player` struct and player statuses.
        *   `ai.go`: Contains the logic for CPU player decisions based on their assigned `AIProfile`.
        *   `betting_limit.go`: Implements the `BettingLimitCalculator` interface (Strategy Pattern) to handle different betting structures like Pot-Limit and No-Limit.
        *   `pot.go`: Manages complex pot calculations, including side pots for all-in situations.
    *   **`util/`**: General-purpose utility functions.
        *   `logger.go`: Initializes and configures the `logrus` logger.

This structure follows the **Separation of Concerns** principle. For example, the `pkg/poker` engine is completely decoupled from the user interface (`internal/cli`), which would allow for replacing the CLI with a web or GUI front-end while reusing the entire game engine and logic.
