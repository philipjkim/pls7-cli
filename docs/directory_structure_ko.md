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
│   └── util/
│       └── logger.go
├── pkg/
│   ├── poker/
│   │   ├── card.go
│   │   ├── deck.go
│   │   ├── evaluation.go
│   │   ├── odds.go
│   │   ├── rules.go
│   │   └── ... (및 테스트 파일)
│   └── engine/
│       ├── action.go
│       ├── ai.go
│       ├── betting_limit.go
│       ├── config.go
│       ├── game.go
│       ├── player.go
│       ├── pot.go
│       ├── run.go
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
    *   `root.go`: 루트 `pls7` 명령어를 생성하고, 플래그를 정의하며, `pkg/engine` 및 `internal/cli`를 호출하여 게임 흐름을 조율하는 메인 게임 루프를 포함합니다.

*   **`rules/`**
    *   다양한 포커 변형의 규칙을 정의하는 YAML 파일을 포함합니다. 이를 통해 애플리케이션이 범용 포커 엔진으로 기능할 수 있습니다.
    *   `nlh.yml`: 노리밋 홀덤(No-Limit Hold'em) 규칙.
    *   `pls.yml`: 팟리밋 삼평(Pot-Limit Sampyeong) 규칙.
    *   `pls7.yml`: 팟리밋 삼평 7-or-Better(Pot-Limit Sampyeong 7-or-Better) 규칙.

*   **`pkg/`**
    *   재사용 가능한 도메인 특화 라이브러리를 포함합니다. 이 디렉토리의 코드는 독립적이며 `internal` 패키지에 대한 의존성이 없습니다. 다른 프로젝트에서 가져다 쓸 수 있습니다.
    *   **`poker/`**: 핵심 포커 라이브러리입니다. 포커의 규칙, 데이터 모델, 평가 로직에 중점을 둔 순수 라이브러리입니다.
        *   `rules.go`: 포커 게임의 속성을 정의하는 계약인 `GameRules` 구조체를 정의합니다.
        *   `card.go`, `deck.go`: 카드와 덱 구조체 및 연산을 정의합니다.
        *   `evaluation.go`: 제공된 `GameRules`에 따라 핸드를 평가합니다.
        *   `odds.go`: 팟 오즈, 에퀴티, 아우츠 계산 로직을 담습니다.
    *   **`engine/`**: 게임 엔진입니다. 포커 게임의 상태와 흐름을 관리합니다.
        *   `game.go`: 실행 중인 게임의 전체 상태를 보유하는 중앙 `Game` 구조체를 정의합니다.
        *   `run.go`: 단일 핸드의 상태 머신(카드 분배, 액션 처리, 페이즈 진행)을 구현합니다.
        *   `player.go`, `pot.go`, `ai.go`: 게임 진행을 위한 핵심 구성 요소와 로직을 정의합니다.
        *   `betting_limit.go`: 다양한 베팅 구조(팟리밋, 노리밋)를 위한 전략을 구현합니다.

*   **`internal/`**
    *   이 CLI 프로젝트에만 해당하는 내부 애플리케이션 코드를 포함합니다. 다른 프로젝트에서 임포트하는 것을 의도하지 않습니다.
    *   **`config/`**: `/rules` 디렉토리의 규칙 파일을 로드하고 `poker.GameRules` 구조체로 파싱하는 역할을 합니다.
    *   **`cli/`**: CLI의 "View"와 "Input" 계층을 관리합니다.
        *   `display.go`: `engine.Game` 상태를 콘솔에 렌더링합니다.
        *   `input.go`: 사용자로부터 액션을 입력받고 파싱합니다.
        *   `format.go`: 출력 포맷팅을 위한 헬퍼 함수를 제공합니다.
    *   **`util/`**: 로거 초기화와 같은 범용 유틸리티 함수.

이 구조는 **관심사의 분리(Separation of Concerns)** 원칙을 따릅니다. 핵심 엔진(`pkg/poker` 및 `pkg/engine`)은 사용자 인터페이스(`internal/cli`)와 완전히 분리되어 있어, 나중에 전체 게임 엔진을 재사용하면서 CLI를 웹 또는 GUI 프론트엔드로 교체할 수 있습니다.
