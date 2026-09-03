# Turbo Go (Version 1.0) 🚀

> **Retro Borland Turbo Pascal / Turbo C Look & Feel IDE for the Go Language**

**Turbo Go**는 90년대 볼랜드(Borland)의 전설적인 **Turbo Pascal 7.0**과 **Turbo C++ 3.0** 특유의 비주얼 인터페이스(Turbo Vision 파란색 에디터 창, 이중선 프레임 `╔═╗`, 상단 드롭다운 메뉴바, 하단 핫키 바, `Alt+F5` User Screen)에 현대의 **Go 컴파일러 및 Delve 디버거**를 완벽하게 연동한 레트로 터미널 개발 환경(TUI IDE)입니다.

---

## 📸 주요 특징

- **Classic Borland Turbo Vision UI**:
  - 시그니처 터보 블루 에디터 캔버스 (`#0000A8`) 및 이중선 박스 드로잉 (`╔═╗`, `║ ║`, `╚═╝`)
  - 입체 텍스트 그림자(Drop Shadow) 및 레트로 윈도우 헤더 (`[■] 1 NONAME00.GO [▲]`)
  - 상단 풀다운 메뉴바 (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Window`, `Help`)
  - 하단 단축키 바 (`F1 Help`, `F2 Save`, `F3 Open`, `Alt+F9 Compile`, `F9 Make`, `Ctrl+F9 Run`, `Alt+F5 User`, `F10 Menu`)
- **Go Syntax Highlighting**:
  - Go 키워드, 타입, 리터럴(문자열, 숫자, 룬), 내장 함수, 주석 구문 강조
- **Compiling 모달 다이얼로그**:
  - 컴파일 시 볼랜드 특유의 "Compiling..." 팝업 박스(대상 파일, 총 라인 수, 에러/워닝 수, 빌드 경과 시간)
  - 컴파일 에러 발생 시 파일명/라인/에러 목록 표시 및 **에디터 해당 라인으로 즉시 점프**
- **Alt+F5 User Screen**:
  - Turbo C의 상징적인 기능! 빌드 실행 결과를 별도의 전체화면 DOS 콘솔 뷰어로 전환하여 확인하고, 아무 키나 누르면 다시 Turbo Go IDE로 복귀
- **인터랙티브 Delve 디버거 연동**:
  - `F4`로 브레이크포인트(`●`) 설정/해제 ➔ **한 줄 전체 빨간색 바(`Red`)** 표시
  - `F5` 디버깅 시작 / Continue, `F8` Step Over, `F7` Trace Into
  - 실행 중 현재 멈춘 라인은 **한 줄 전체 노란색 바(`Yellow`)**로 시선 집중
  - 하단 **Watches 윈도우(`Alt+W`)**를 통해 로컬 변수명, 타입, 값을 실시간 감시
- **볼랜드 레트로 사운드 이펙트 (Sound FX)**:
  - 컴파일 성공 시 경쾌한 2단 비프음, 컴파일 에러 시 묵직한 에러음
  - 디버거 브레이크포인트 적중 및 스텝 이동 시 기분 좋은 핑 사운드
  - `F10` ➔ `Options` ➔ `Sound: ON / OFF` 토글 지원

---

## ⌨️ 단축키 안내

| 단축키 | 기능 | 설명 |
|---|---|---|
| **F1** | Help / About | Turbo Go 정보 및 도움말 대화상자 |
| **F2** | Save | 현재 버퍼 저장 / 다른 이름으로 저장 |
| **F3** | Open | 파일 브라우저 다이얼로그 열기 |
| **F4** | **Breakpoint** | 현재 라인 브레이크포인트(`●`) 설정/해제 |
| **F5** | **Debug / Continue** | 디버깅 시작 / 다음 브레이크포인트까지 계속 실행 |
| **F7** | **Trace Into** | 한 줄씩 실행 (함수 내부 진입) |
| **F8** | **Step Over** | 한 줄씩 실행 (함수 건너뛰기) |
| **Ctrl + F2** | **Reset Debugger** | 디버깅 세션 종료 및 실행 포인터 초기화 |
| **Alt + W** | **Watches Window** | 하단 변수 감시(Watches) 윈도우 토글 |
| **Ctrl + F** | **Find** | 문자열 검색 다이얼로그 열기 |
| **Ctrl + L** | **Search Again** | 이전 검색어로 다음 위치 계속 찾기 (Find Next) |
| **Alt + G** | **Go to Line** | 특정 줄 번호로 커서 이동 (`Ctrl+G`) |
| **Alt + L** | **Line Numbers** | 왼쪽 줄 번호 표시 On / Off 토글 (`Option+L`, `F6`) |
| **Ctrl + F9** | **Run** | 빌드 후 일반 실행 및 결과 화면(**User Screen**) 표시 |
| **Alt + F9** | **Compile** | "Compiling..." 통계 팝업과 함께 빌드 실행 |
| **F9** | Make | 빌드 실행 |
| **Alt + F5** | **User Screen** | 프로그램 실행 결과 화면 토글 |
| **F10** | Menu Bar | 상단 풀다운 메뉴바 포커스 토글 |
| **Alt + X** | Exit | Turbo Go 종료 |
| **Shift + 방향키** | **Select Block** | 텍스트 영역 블록 선택 (하이라이트) |
| **Ctrl + Ins** | **Copy** | 선택한 블록 클립보드에 복사 (`Edit ➔ Copy`) |
| **Shift + Del** | **Cut** | 선택한 블록 잘라내기 (`Edit ➔ Cut`) |
| **Shift + Ins** | **Paste** | 클립보드 내용 커서 위치에 붙여넣기 (`Edit ➔ Paste`) |
| **Esc** | Close | 활성 메뉴/팝업 다이얼로그 닫기, 선택 해제 |

---

## 🛠️ 실행 및 빌드 방법

### 1. 바이너리 빌드 및 실행
```bash
go build -o bin/tg ./cmd/tg
./bin/tg
```

### 2. 특정 Go 파일 열기
```bash
./bin/tg examples/hello.go
```
