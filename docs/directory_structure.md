### 📂 프로젝트 구조 (현재)

```
pls7-cli/
├── cmd/
│   ├── root.go
│   └── play.go
├── internal/
│   ├── cli/
│   │   ├── display.go
│   │   └── input.go
│   ├── game/
│   │   ├── action.go
│   │   ├── ai.go
│   │   ├── ai_test.go
│   │   ├── betting.go
│   │   ├── betting_test.go
│   │   ├── config.go
│   │   ├── game.go
│   │   ├── game_test.go
│   │   ├── player.go
│   │   ├── pot.go
│   │   ├── pot_test.go
│   │   └── run.go
│   └── util/
│       ├── format.go
│       ├── format_test.go
│       └── logger.go
├── pkg/
│   └── poker/
│       ├── card.go
│       ├── deck.go
│       ├── deck_test.go
│       ├── evaluation.go
│       ├── evaluation_test.go
│       ├── odds.go
│       └── odds_test.go
├── main.go
└── go.mod
```

### 각 디렉토리 및 패키지의 역할

* **`main.go`**

    * 프로그램의 가장 최상위 진입점입니다.
    * 오직 `cmd.Execute()` 함수 하나만 호출하는 단순한 역할을 합니다.

* **`cmd/`**

    * 모든 CLI 명령어와 플래그를 정의하고 관리합니다.
    * `root.go`: `pls7` 이라는 루트 명령어를 생성합니다.
    * `play.go`: 실제 게임을 시작하는 `play` 하위 명령어를 정의하고, `--difficulty`, `--dev` 등 게임 시작 옵션(플래그)을 처리합니다.

* **`pkg/`**

    * **재사용 가능한 순수 로직**을 담는 패키지입니다. 이 프로젝트의 범위를 넘어 다른 프로젝트에서도 이론적으로 가져다 쓸 수 있는 코드입니다.
    * **`poker/`**: 게임의 핵심 규칙 로직을 담습니다. **CLI나 게임 흐름에 대한 어떤 정보도 알지 못하는** 순수한 상태여야 합니다.
        * `card.go`: `Card`, `Suit`, `Rank` 구조체 및 관련 함수 정의.
        * `deck.go`: `Deck` 구조체와 셔플, 딜링 기능.
        * `deck_test.go`: `deck.go`에 대한 테스트 코드.
        * `evaluation.go`: 가장 복잡한 **족보 판정 로직** 전체.
        * `evaluation_test.go`: `evaluation.go`에 대한 테스트 코드.
        * `odds.go`: 팟 오즈(Pot Odds), 에퀴티(Equity), 아우츠(Outs) 등 확률 계산 관련 로직.
        * `odds_test.go`: `odds.go`에 대한 테스트 코드.

* **`internal/`**

    * **이 애플리케이션에만 종속적인** 내부 로직을 담습니다. `pkg/`와 달리 다른 프로젝트에서 임포트하여 사용하는 것을 의도하지 않습니다.
    * **`cli/`**: CLI의 **View(화면)** 와 **Input(입력)** 을 전담합니다.
        * `display.go`: 게임 보드, 플레이어 상태 등 모든 것을 화면에 예쁘게 그려주는 함수들을 모아둡니다.
        * `input.go`: 사용자로부터 액션(`c`, `r`, `f` 등)을 입력받고 파싱하는 로직을 담당합니다.
    * **`game/`**: 게임의 전체 흐름을 지휘하는 **오케스트레이터**입니다. `cli`, `poker`, `util` 등 다른 패키지의 기능을 가져와 조립합니다.
        * `game.go`: `Game` 구조체를 정의하고, 게임의 전반적인 상태(페이즈, 팟, 플레이어 등)를 관리합니다.
        * `game_test.go`: `game.go`의 핵심 로직에 대한 테스트 코드.
        * `run.go`: `StartNewHand`, `ProcessAction`, `ExecuteBettingLoop` 등 실제 게임 한 판의 시작, 진행, 종료와 관련된 함수들을 포함합니다.
        * `player.go`: `Player` 구조체와 플레이어의 상태(`Playing`, `Folded` 등)를 정의합니다.
        * `action.go`: `PlayerAction` 구조체와 플레이어가 할 수 있는 액션의 종류(`Fold`, `Check` 등)를 정의합니다.
        * `betting.go`: `CalculateBettingLimits` 등 베팅과 관련된 복잡한 규칙(팟 리밋 등)을 계산하는 로직을 담습니다.
        * `betting_test.go`: 다양한 베팅 시나리오에 대한 테스트 코드.
        * `pot.go`: `DistributePot` 등 메인 팟과 사이드 팟을 계산하고 분배하는 로직을 담당합니다.
        * `pot_test.go`: 복잡한 팟 분배 시나리오(사이드 팟 등)에 대한 테스트 코드.
        * `ai.go`: `getEasyAction`, `getMediumAction` 등 CPU 플레이어의 행동 로직을 담습니다.
        * `ai_test.go`: CPU AI 로직에 대한 테스트 코드.
        * `config.go`: 블라인드 금액, AI 난이도 상수 등 게임의 기본 설정 값을 정의합니다.
    * **`util/`**: 특정 도메인에 종속되지 않는 범용 유틸리티 함수들을 모아둡니다.
        * `format.go`: 숫자 포맷팅 등 문자열 форматирование 관련 함수.
        * `format_test.go`: `format.go`에 대한 테스트 코드.
        * `logger.go`: `logrus`를 사용한 로거 초기화 및 설정.

이 구조는 **관심사의 분리(Separation of Concerns)** 원칙을 잘 따르고 있어, 예를 들어 나중에 CLI가 아닌 웹 UI로 바꾸고 싶을 때 `internal/cli/`만 교체하고 `pkg/poker/`와 `internal/game/`의 대부분은 그대로 재사용할 수 있게 됩니다.