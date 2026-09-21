# 90년대 볼랜드 터보 파스칼 & C++ IDE의 감성이 그리워, Go, Rust, FORTRAN 77용 네이티브 Turbo TUI IDE를 만들었습니다

안녕하세요 개발자 여러분!

80년대 후반이나 90년대에 프로그래밍을 배우셨다면, **볼랜드 터보 파스칼 7.0 (Borland Turbo Pascal 7.0)**이나 **터보 C++ 3.0 (Turbo C++ 3.0)**의 마법 같은 경험을 기억하실 겁니다:
- 짙은 파란색 배경과 선명한 노란색 텍스트 및 메뉴
- `Ctrl+F9`를 누르자마자 1초도 안 걸려 즉시 완료되는 컴파일
- 텍스트 창에서 매끄럽게 작동하던 풀다운 메뉴, 대화 상자, 마우스 지원
- 실행하자마자 0.1초 만에 눈앞에 나타나는 개발 환경

VS Code 같은 최신 IDE들도 정말 훌륭하지만, 원격 클라우드 VM, 도커 컨테이너, 헤드리스 슈퍼컴퓨터 클러스터에 SSH로 접속해 작업할 때마다 매번 마주치는 불편함들이 있었습니다:
- **VS Code Remote-SSH**는 메모리를 잡아먹는 무거운 Node.js 서버를 백그라운드에 띄워서, 1GB RAM 수준의 저사양 인스턴스에서는 종종 OOM(메모리 부족)으로 다운되곤 합니다.
- **Neovim**도 훌륭하지만, 수십 대의 원격 장비마다 수십 개의 플러그인과 복잡한 Lua 설정을 관리하는 것은 피곤한 일입니다.
- **Nano / Micro** 같은 기본 에디터는 여러 파일을 넘나들거나, 심볼 정의로 이동하고 컴파일러 에러를 확인하기에는 기능이 너무 부족합니다.

그래서 지난 몇 달간 그 시절 볼랜드 터보 시리즈의 번개처럼 빠르고 향수 어린 경험을 되살려, 최신 및 레거시 언어를 위한 **의존성 없는 독립형 TUI IDE 시리즈(Turbo TUI Trilogy)**를 직접 구현했습니다:

---

### 3가지 프로젝트 소개:

#### 1. `tg` — Turbo Go
* **탄생 배경**: Go 언어는 설계 철학과 계보 면에서 파스칼/모듈라의 DNA를 직접 공유하고 있습니다.
* **매력 포인트**: Go의 컴파일러는 매우 빠른 것으로 유명합니다. `tg`에서 컴파일 버튼을 누르면 터보 파스칼의 "0.1초 만에 빌드되는" 즉각적인 쾌감을 그대로 느낄 수 있습니다.
* **추천 용도**: CLI 도구 작성, 백엔드 마이크로서비스 개발, SSH 접속 환경에서 딜레이 없는 가벼운 원격 개발.

#### 2. `tr` — Turbo Rust
* **탄생 배경**: 터보 C++의 정통 감성과 현대적인 시스템 프로그래밍 언어의 만남.
* **매력 포인트**: `cargo check` 및 `cargo build`와 직접 연동됩니다. 러스트 컴파일러가 에러를 뱉으면, `tr`이 이를 파싱하여 고전적인 볼랜드 모달 에러 목록 창을 띄워줍니다. 목록에서 `Enter`만 누르면 해당 파일, 줄 번호, 컬럼 위치로 즉시 점프합니다.
* **추천 용도**: 시스템 프로그래밍, Rust 언어 학습, 알고리즘 문제 풀이 등 잡음 없는 몰입 환경이 필요할 때.

#### 3. `tf77` — Turbo FORTRAN 77
* **탄생 배경**: 고성능 컴퓨팅(HPC), 물리, 과학 시뮬레이션 연구자들을 위한 헌정 프로젝트.
* **매력 포인트**: 기상 예측, 항공우주 유체역학(CFD), 분자역학, 원자력 등 검증된 수많은 핵심 과학 코드 수백만 줄이 여전히 GUI 없는 슈퍼컴퓨터에서 고정 형식(Fixed-form) FORTRAN 77로 돌아가고 있습니다. 현대적인 에디터들은 F77의 고정 형식을 다루기에 매우 어색합니다.
* **F77 특화 기능**:
  - **72열 제한 가이드라인** (72열을 넘어가는 코드는 컴파일러가 무시하거나 에러 발생).
  - **6열 줄 연속(Continuation) 문자 표시선**.
  - 프로젝트 전체에 걸친 대소문자 무시 `SUBROUTINE`, `FUNCTION`, `PROGRAM` 정의 점프.
* **추천 용도**: VS Code 사용이 정책적으로 금지되거나 무거워서 원격 클러스터 로그인 노드에서 레거시 코드를 직접 수정해야 하는 대학원생, 물리학자, 엔지니어.

---

### 세 IDE의 공통 핵심 기능:

* **정통 볼랜드 Turbo Vision UI**:
  - 시그니처 터보 블루 화면 (`#0000A8`), 이중 테두리 박스 프레임 (`╔═╗`), 단축키 강조(Mnemonic) 풀다운 메뉴, 입체 그림자, 빌드 성공/실패 시 정통 PC 스피커 사운드 이펙트.
* **인터랙티브 디버거 & 실시간 변수 감시 (Watches Window)**:
  - `F4`로 브레이크포인트(`●`) 설정/해제, `F5` 디버깅 시작/Continue, `F8` Step Over, `F7` Trace Into.
  - 현재 실행 라인을 시각적으로 한눈에 포착하는 노란색 강조 바.
  - 하단 **Watches Window**를 통해 로컬 변수명, 타입, 값을 실시간 감시 (Go: Delve, Rust: GDB/LLDB, Fortran: 네이티브 F77 인터프리터 및 LLDB 지원).
* **정의로 이동 (`F12` / Go to Definition)**:
  - 함수, 구조체, 타입, 서브루틴 위에 커서를 두고 `F12`를 누르면, 멀티 파일 프로젝트 전체를 탐색하여 해당 정의 위치로 즉시 이동합니다(다른 파일에 있다면 자동 로드).
* **Alt+F5 User Screen**:
  - Turbo C의 상징적인 기능! 프로그램 실행 결과를 별도의 전체화면 콘솔 화면으로 전환하여 확인하고, 아무 키나 누르면 다시 IDE로 복귀.
* **현대적인 조작성 및 단어 단위 이동**:
  - 메뉴 드롭다운에서 알파벳 한 글자로 즉시 실행하는 서브메뉴 핫키 (예: `File` ➔ `N` New, `O` Open, `S` Save, `A` Save As).
  - 텍스트 작성 중 `Ctrl+Left/Right` 및 macOS `Option+Left/Right`를 통한 단어 단위 고속 이동 및 블록 선택(`Shift` 조합).
* **외부 의존성 제로 (초경량 단일 정적 바이너리)**:
  - 약 10~15MB 크기의 단일 정적 Go 바이너리로 빌드됩니다. Node.js, Python, Electron 같은 런타임이 전혀 필요 없으며, 원격 서버나 저사양 VM에 파일 하나만 올리면 0.01초 만에 실행됩니다.

---

### 빠른 설치 및 실행

#### 1. 사전 빌드된 바이너리 다운로드 (GitHub Releases)
Releases 페이지에서 **macOS (Apple Silicon)**, **Linux (x86_64)**, **Windows (x64)**용 단일 실행 파일을 즉시 다운로드하실 수 있습니다:
- [Turbo Go 릴리즈 (v0.90)](https://github.com/MaroShim/TurboGo/releases/latest)
- [Turbo Rust 릴리즈 (v0.90)](https://github.com/MaroShim/TurboRust/releases/latest)
- [Turbo Fortran 릴리즈 (v0.90)](https://github.com/MaroShim/TurboF77/releases/latest)

#### 2. `go install`을 통해 간편하게 설치:

```bash
# Turbo Go
go install github.com/MaroShim/tg/cmd/tg@latest

# Turbo Rust
go install github.com/MaroShim/TurboRust/cmd/tr@latest

# Turbo Fortran (클래식 F77 및 현대 F90+ 모두 지원)
go install github.com/MaroShim/tf77/cmd/tf77@latest
go install github.com/MaroShim/tf77/cmd/tf@latest
```

#### 3. 소스 코드에서 직접 빌드:

```bash
git clone https://github.com/MaroShim/TurboGo.git && cd TurboGo && go build -o bin/tg ./cmd/tg
```

---

### 저장소 링크

- **Turbo Go**: [github.com/MaroShim/TurboGo](https://github.com/MaroShim/TurboGo)
- **Turbo Rust**: [github.com/MaroShim/TurboRust](https://github.com/MaroShim/TurboRust)
- **Turbo Fortran**: [github.com/MaroShim/TurboF77](https://github.com/MaroShim/TurboF77)

*(게시 시 파란색 에디터 화면 스크린샷 1~2장을 함께 첨부하면 반응이 훨씬 좋습니다!)*

---

예전 볼랜드 IDE로 코딩하던 시절의 추억이나, 지금도 슈퍼컴퓨터에서 포트란과 씨름하고 계신 연구원분들의 피드백을 언제든 환영합니다! 다음에 추가되었으면 하는 기능이나 지원했으면 하는 언어가 있다면 자유롭게 의견 남겨주세요!
