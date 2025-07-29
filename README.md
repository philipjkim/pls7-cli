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
# standard
go run main.go play

# for debugging (not clearing previous output)
go run main.go play --dev
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

## 📝 Development Plan (Step-by-Step)

### **1단계: 프로젝트 뼈대 구축 (Cobra & 디렉토리)**

* **목표**: 명령어로 실행 가능한 가장 기본적인 "Hello World" 앱 만들기.
* **결과물**: `go run main.go play` 명령어를 터미널에 입력하면 환영 메시지가 출력되는, 뼈대만 갖춘 CLI 애플리케이션.
* **완료여부**: ✅

---

### **2단계: 기본 데이터 모델 구현 (Card & Deck)**

* **목표**: 카드와 덱을 코드로 구현하고, 셔플된 덱을 화면에 출력하기.
* **결과물**: 실행할 때마다 무작위로 섞인 카드가 화면에 출력되는 앱.
* **완료여부**: ✅

---

### **3단계: 정적(Static) 핸드 시뮬레이션**

* **목표**: 게임 흐름 없이, 카드 분배가 완료된 한순간의 테이블 상황을 화면에 보여주기.
* **결과물**: 6명의 플레이어 패와 커뮤니티 카드가 모두 깔린, 한 판의 정지된 스냅샷이 화면에 출력되는 앱.
* **완료여부**: ✅

---

### **4단계: 족보 판정 로직 구현 및 테스트**

* **목표**: 가장 복잡한 족보 판정 로직을 단위 테스트를 통해 독립적으로 완성하기.
* **결과물**: 정적 핸드 상황에서 각 플레이어의 하이/로우 핸드 족보를 정확히 계산하여 알려주는 앱.
* **완료여부**: ✅

---

### **5단계: 게임 흐름 및 상태 관리 구현**

* **목표**: 실제 게임 한 판의 흐름(라운드, 베팅, 플레이어 상태 등)을 관리하는 데이터 구조와 자동 진행 로직의 뼈대를 만들기.
* **주요 작업**:
  1.  `Player`와 `Game` 구조체에 칩, 상태, 팟, 페이즈 등 동적 필드 추가.
  2.  순환 참조를 피해, `cmd/play.go`가 지휘하는 자동 게임 루프(Pre-flop -> Flop -> Turn -> River) 구현.
  3.  실제 블라인드(SB/BB) 베팅 로직 구현 및 정확한 팟 계산 확인.
* **결과물**: 사용자 입력 없이 프리플랍부터 리버까지 카드 깔리는 과정이 순차적으로 보이고, 블라인드 팟이 정확히 계산되는 '시네마틱 모드' 게임.
* **완료여부**: ✅

---

### **6단계: 상호작용 가능한 베팅 라운드 구현**

* **목표**: 플레이어가 직접 자신의 턴에 명령어를 입력하고, 게임이 그에 반응하게 하기.
* **주요 작업**:
  1.  `PlayerAction` 데이터 구조 정의.
  2.  사용자 입력을 받는 프롬프트(`internal/cli/input.go`) 구현.
  3.  `Fold`, `Check`, `Call` 액션 처리 로직 구현.
  4.  `cmd/play.go`의 메인 루프가 플레이어 턴에 멈추고 입력을 기다리도록 수정.
* **결과물**: 사용자가 직접 폴드, 체크, 콜을 하며 참여할 수 있는, 최초의 **플레이 가능한 버전**.
* **완료여부**: ✅

---

### **7단계: 고급 베팅 로직 및 팟 리밋 구현**

* **목표**: 실제 포커 게임처럼 베팅과 레이즈가 오고 가는 완전한 베팅 라운드를 구현.
* **주요 작업**:
  * 사용자로부터 `Bet`, `Raise` 금액을 입력받는 로직 추가.
  * 팟 리밋(Pot-Limit) 규칙에 따라 베팅/레이즈 가능한 최소/최대 금액 계산 로직 구현.
  * 한 명의 베팅/레이즈 이후, 다른 모든 플레이어에게 다시 액션 기회가 돌아가는 완전한 베팅 루프 로직 구현.
* **완료여부**: ✅

---

### **8단계: 전체 기능 완성 및 CPU AI 구현**

* **목표**: 나머지 기능들을 추가하고 CPU AI를 구현하여 완전한 싱글 플레이어 게임으로 발전시키기.
* **주요 작업**:
  * 팟 분배(Pot Distribution) 로직 구현.
  * 게임 오버 및 다음 핸드 시작 로직 구현.
  * 정의했던 난이도별(Easy, Medium, Hard) CPU AI 로직 구현.
* **완료여부**: ✅

---

### **9단계: 최종 폴리싱 및 리팩토링**

* **목표**: 전체 코드를 정리하고, 테스트를 추가하여 배포 준비 완료.
* **주요 작업**:
  * 개발 모드 및 로깅 시스템 도입. (✅)
  * 조건부 화면 클리어 기능 구현. (✅)
  * 정확한 팟 리밋(Pot-Limit) 계산 로직. (✅)
  * 베팅 루프 안정성 강화를 위한 테스트 추가.
* **완료여부**: ⏳ (현재 진행 중)
