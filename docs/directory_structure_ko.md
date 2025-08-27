# 디렉토리 구조

이 문서는 `pls7-cli` 프로젝트의 디렉토리 구조를 설명합니다.

```
pls7-cli/
├── cmd/
│   └── root.go
├── docs/
│   ├── architecture.md
│   ├── development_plan.md
│   ├── directory_structure.md
│   └── ... (기타 문서)
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
│   │   └── ... (및 테스트 파일)
│   └── util/
│       └── logger.go
├── pkg/
│   └── poker/
│       ├── card.go
│       ├── deck.go
│       ├── evaluation.go
│       ├── odds.go
│       ├── rules.go
│       └── ... (및 테스트 파일)
├── rules/
│   ├── nlh.yml
│   ├── pls.yml
│   └── pls7.yml
├── main.go
├── go.mod
└── README.md
```

### 각 디렉토리 및 패키지의 역할

*   **`main.go`**
    *   프로그램의 메인 진입점입니다.
    *   `cmd.Execute()`를 호출하는 유일한 역할을 합니다.

*   **`cmd/`**
    *   모든 CLI(Command Line Interface) 명령어와 플래그를 정의하고 관리합니다.
    *   `root.go`: 루트 명령어인 `pls7`을 생성하고, 모든 플래그(`--rule`, `--difficulty` 등)를 정의하며, 전체 게임 흐름을 조율하는 메인 게임 루프를 포함합니다.

*   **`rules/`**
    *   다양한 포커 변형의 규칙을 정의하는 YAML 파일을 포함합니다. 이를 통해 애플리케이션이 범용 포커 엔진으로 기능할 수 있습니다.
    *   `nlh.yml`: 노리밋 홀덤(No-Limit Hold'em) 규칙.
    *   `pls.yml`: 팟리밋 삼평(Pot-Limit Sampyeong) 규칙.
    *   `pls7.yml`: 팟리밋 삼평 7-or-Better(Pot-Limit Sampyeong 7-or-Better) 규칙.

*   **`pkg/`**
    *   재사용 가능한 도메인 특화 라이브러리를 포함합니다. 이 디렉토리의 코드는 독립적이며 `internal` 패키지에 대한 의존성이 없습니다. 이론적으로 다른 프로젝트에서 가져다 쓸 수 있습니다.
    *   **`poker/`**: 핵심 포커 엔진입니다. CLI나 특정 게임 흐름에 대해 전혀 알지 못하는 순수 라이브러리입니다.
        *   `rules.go`: YAML 파일로부터 채워지는 `GameRules` 구조체를 정의합니다. 이는 포커 게임의 속성을 정의하는 계약과 같습니다.
        *   `card.go`, `deck.go`: 카드와 덱 구조체 및 기본 연산을 정의합니다.
        *   `evaluation.go`: 엔진의 가장 복잡한 부분입니다. 제공된 `GameRules`에 따라 핸드를 평가합니다(예: 표준 족보, 스킵 스트레이트, 로우 핸드).
        *   `odds.go`: 팟 오즈, 에퀴티, 아우츠 계산 로직을 담습니다.

*   **`internal/`**
    *   이 프로젝트에만 해당하는 모든 내부 애플리케이션 코드를 포함합니다. 다른 프로젝트에서 임포트하는 것을 의도하지 않습니다.
    *   **`config/`**: `/rules` 디렉토리의 규칙 파일을 로드하고 파싱하는 역할을 합니다.
        *   `rules.go`: YAML 파일을 읽고 `poker.GameRules` 구조체로 변환하는 로직을 포함합니다.
    *   **`cli/`**: CLI의 "View"와 "Input" 계층을 관리합니다.
        *   `display.go`: 게임 상태(보드, 플레이어, 팟)를 콘솔에 렌더링합니다.
        *   `input.go`: 사용자로부터 액션(체크, 벳, 레이즈, 폴드)을 입력받고 파싱합니다.
        *   `format.go`: 숫자에 쉼표를 추가하는 등 출력 포맷팅을 위한 헬퍼 함수를 제공합니다.
    *   **`game/`**: 애플리케이션의 오케스트레이터입니다. `poker` 엔진과 `cli`를 연결하고 게임 상태와 턴 기반 흐름을 관리합니다.
        *   `game.go`: 실행 중인 게임의 전체 상태(플레이어, 덱, 페이즈, 팟 등)를 보유하는 중앙 `Game` 구조체를 정의합니다.
        *   `run.go`: 단일 핸드의 상태 머신(카드 분배, 액션 처리, 페이즈 진행)을 구현합니다.
        *   `player.go`: `Player` 구조체와 플레이어 상태를 정의합니다.
        *   `ai.go`: 할당된 `AIProfile`에 따라 CPU 플레이어의 의사 결정을 위한 로직을 포함합니다.
        *   `betting_limit.go`: 팟리밋, 노리밋과 같은 다양한 베팅 구조를 처리하기 위해 `BettingLimitCalculator` 인터페이스(전략 패턴)를 구현합니다.
        *   `pot.go`: 올인 상황에서의 사이드 팟을 포함한 복잡한 팟 계산을 관리합니다.
        *   `turn.go`: 활성 플레이어 간의 턴을 진행하는 로직.
    *   **`util/`**: 범용 유틸리티 함수.
        *   `logger.go`: `logrus` 로거를 초기화하고 설정합니다.

이 구조는 **관심사의 분리(Separation of Concerns)** 원칙을 따릅니다. 예를 들어, `pkg/poker` 엔진은 사용자 인터페이스(`internal/cli`)와 완전히 분리되어 있어, 나중에 전체 게임 엔진과 로직을 재사용하면서 CLI를 웹 또는 GUI 프론트엔드로 교체할 수 있습니다.
