# Turbo Go (`tg`) 프로젝트 구축 및 워크스루 🚀

볼랜드 Turbo Vision TUI 아키텍처와 감성을 완벽하게 계승하면서, **Go 언어 툴체인(`go build`)과 Delve(`dlv`) 인터랙티브 디버거**를 전면 연동한 레트로 TUI IDE **Turbo Go (`tg`)** 종합 워크스루 문서입니다.

---

## 📸 프로젝트 디렉터리 구조 및 아키텍처

```
/Users/maro/Projects/tg/
├── go.mod                      # Go 모듈 정의 (module tg)
├── go.sum                      # 의존성 체크섬
├── README.md                   # Turbo Go 공식 안내서
├── WALKTHROUGH.md              # 종합 프로젝트 워크스루 문서
├── bin/
│   ├── tg                      # macOS 실행 바이너리
│   └── tg.exe                  # Windows 실행 바이너리
├── cmd/
│   └── tg/
│       └── main.go             # IDE 진입점, 이벤트 루프, 액션 디스패처, 글로벌 단축키
├── examples/
│   └── hello.go                # 기본 Go 예제 코드
└── internal/
    ├── compiler/
    │   ├── runner.go           # go build, multi-file 모듈 감지, 라인 카운터, 에러 파서
    │   └── runner_test.go      # 단일/다중 파일 및 go.mod 빌드 통합 테스트
    ├── debugger/
    │   ├── debugger.go         # Delve(dlv) JSON-RPC v2 양방향 디버거 클라이언트
    │   └── debugger_test.go    # Delve 브레이크포인트, 스텝 실행 테스트
    ├── sound/
    │   └── sound.go            # 2.5인치 IBM 종이 콘 PC 스피커 물리 모델 사운드 합성기
    ├── syntax/
    │   ├── go_highlighter.go   # Go 키워드, 타입, 리터럴, 주석 구문 강조기
    │   └── go_highlighter_test.go
    └── ui/
        ├── app.go              # UI 애플리케이션 상태 컨트롤러
        ├── clipboard.go        # 클립보드 공통 인터페이스
        ├── clipboard_darwin.go # macOS 전용 클립보드 (pbcopy / pbpaste)
        ├── clipboard_windows.go# Windows Win32 API 클립보드 (user32.dll / kernel32.dll)
        ├── clipboard_other.go  # Linux X11/Wayland 클립보드 (xclip / wl-copy)
        ├── editor.go           # 터보 블루 에디터 버퍼 (블록 선택, 검색 하이라이트)
        ├── editor_test.go      # 에디터 조작, 검색, 클립보드 테스트
        ├── menubar.go          # 상단 볼랜드 풀다운 메뉴바
        ├── statusbar.go        # 하단 단축키 바
        ├── theme.go            # 볼랜드 색상 팔레트 및 이중선 박스 드로잉 기호
        ├── userscreen.go       # Alt+F5 전체화면 DOS 콘솔 뷰어
        ├── watchwindow.go      # Alt+W 변수 감시(Watches) 윈도우
        ├── window.go           # 프레임, 다이얼로그 박스, 그림자 드로잉
        └── dialogs/
            ├── about.go        # Turbo Go 정보 대화상자
            ├── compile.go      # "Compiling..." 통계 팝업 모달
            ├── errorlist.go    # 컴파일 에러 목록 및 에디터 즉시 점프
            ├── find.go         # 문자열 검색 대화상자
            ├── gotoline.go     # 특정 줄 번호 이동
            ├── openfile.go     # .go 필터링 파일 브라우저
            └── savefile.go     # 파일 저장 대화상자
```

---

## 🛠️ 주요 기능 상세

### 1. Classic Borland Turbo Vision UI
- **시그니처 블루 캔버스 (`#0000A8`)**와 정교한 이중선 프레임 (`╔═╗`, `║ ║`, `╚═╝`)
- **텍스트 드롭 섀도우** 및 오리지널 볼랜드 윈도우 헤더 (`[■] 1 NONAME00.GO [▲]`)
- 상단 풀다운 메뉴바 (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Options`, `Window`, `Help`)
- 하단 핫키 상태바 (`F1 Help  F2 Save  F3 Open  Alt+F9 Compile  F9 Make  Ctrl+F9 Run  Alt+F5 User  F10 Menu`)

### 2. Go 구문 강조 & 레트로 에디터 버퍼
- Go 언어 키워드, 빌트인 타입, 문자열 리터럴, 주석 구문 분석 및 레트로 컬러링
- **탭 들여쓰기 디폴트 4 적용**: 탭 스톱(4칸 단위) 자동 정렬 및 스마트 백스페이스
- **기존 탭 문자 자동 확장 (`ExpandTabs`)**: 기존 Go 파일의 `\t` 들여쓰기를 4칸 단위로 깔끔하게 렌더링
- **줄 번호 온/오프 토글 (`Alt+L`, `Option+L`, `F6`)**: 디폴트 Off로 볼랜드 순정 스타일 재현

### 3. Delve 인터랙티브 디버거 & Watches 윈도우
- **F4 브레이크포인트 설정**: 소스코드 거터부터 라인 끝까지 전체 **솔리드 레드 바**로 명확히 표시
- **파일별 브레이크포인트 격리 및 복원**: 여러 파일을 전환해도 파일별 브레이크포인트가 안전하게 보존
- **F5 디버그 시작 / F8 Step Over / F7 Trace Into**: 한 줄씩 단계별 디버깅
- **솔리드 옐로우 IP 바 (`►`)**: 현재 실행 포인터 라인을 황금빛 노란색 바 + 검정 볼드 텍스트로 하이라이트
- **Watches 윈도우 (`Alt+W`)**: 현재 스코프의 로컬 변수 이름, 타입, 값을 실시간 테이블로 출력

### 4. Alt+F5 User Screen (풀스크린 콘솔)
- 프로그램 실행(`Ctrl+F9`) 결과 화면을 가상 콘솔 버퍼에 보존
- **`Alt+F5`**를 누르면 전체 화면 DOS 콘솔 뷰어로 전환되어 실행 결과를 확인하고, 아무 키나 누르면 IDE로 복귀

### 5. 아날로그 IBM PC 콘 스피커 사운드 (Sound FX)
- 1990년대 정품 IBM PC 본체 내부 2.5인치 종이 콘 스피커 물리 모델 시뮬레이션
- 1차 IIR 로우패스 필터(~1400Hz)와 소프트 어택/디케이 엔벨로프를 통한 뭉툭하고 포근한 아날로그 톤
  - **컴파일 성공 (`Alt+F9`)**: 740Hz ➔ 1108Hz의 포근한 2단 비프 ("뾱-뽁!")
  - **컴파일 실패 / 에러**: 196Hz의 묵직하고 둥근 저음 버저 ("동-")
  - **디버깅 스텝 (`F8`/`F7`)**: 880Hz의 짧고 깔끔한 아날로그 탭 ("똑!")
- `Options ➔ Sound: ON / OFF` 토글 지원

### 6. 문자열 검색 (`Ctrl+F`, `Ctrl+L`) & 시각화
- **`Ctrl+F`**: 볼랜드 스타일의 `Find` 모달 다이얼로그 (대소문자 구분 체크박스 지원)
- **`Ctrl+L`**: 다음 일치 단어로 계속 점프 (Search Again / Find Next)
- **청록색 매칭 하이라이트**: 찾은 단어 전체를 **선명한 청록색(`Light Cyan`) 블록**으로 반전 표시하고 화면 중앙으로 자동 스크롤

### 7. 텍스트 블록 선택 및 복사 / 잘라내기 / 붙여넣기
- **`Shift + 방향키`**: 텍스트 영역을 청록색 블록으로 자유롭게 긁어 선택
- **`Ctrl + Insert` (`Edit ➔ Copy`)**: 선택 블록을 내부 클립보드 및 OS 시스템 클립보드에 동시 복사
- **`Shift + Delete` (`Edit ➔ Cut`)**: 선택 블록 잘라내기
- **`Shift + Insert` (`Edit ➔ Paste`)**: 커서 위치에 클립보드 내용 붙여넣기

### 8. 다중 파일 및 Go 모듈 프로젝트 컴파일 지원
- **`go.mod` 모듈 프로젝트 자동 감지 (`FindGoModuleRoot`)**: 상위 폴더의 `go.mod`를 찾아 패키지 단위(`go build -o tmpBin .`)로 빌드하여 모듈 임포트와 다중 소스 파일 완벽 컴파일
- **`go.mod` 없는 동일 디렉터리 다중 소스 파일 자동 수집**: 같은 디렉터리의 동일 패키지 소스 파일들을 일괄 수집하여 컴파일하므로 `undefined` 에러 방지
- **에러 발생 파일 자동 열기**: 다른 파일에서 컴파일 에러 발생 시 에디터가 해당 소스 파일을 자동으로 열고(`editor.LoadFile`) 해당 위치로 즉시 점프
- **정확한 총 라인 수 집계 (`CountLines`)**: 프로젝트 내 모든 관련 `.go` 파일 줄 수를 합산하여 통계 팝업에 표시

### 9. macOS & Windows 완벽 크로스 플랫폼 호환
- Go 표준 조건부 빌드 태그(`//go:build`)를 적용하여 클립보드 및 사운드, 파일 경로 완벽 분리
- Windows의 `user32.dll` / `kernel32.dll` Win32 네이티브 클립보드 API 및 `.exe` 바이너리 지원
- Windows 드라이브 문자(`C:\...`) 정규식 에러 파서 지원

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
| **Ctrl + F2** | **Reset Debugger**| 디버깅 세션 종료 및 실행 포인터 초기화 |
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

## 🧪 검증 결과

1. **단위 테스트**:
   - `go test -count=1 ./...` ➔ 모든 테스트 패키지(`internal/compiler`, `internal/debugger`, `internal/syntax`, `internal/ui`) **100% PASS**
2. **다중 파일 및 모듈 빌드 테스트**:
   - `TestCompilerMultiFileBuild` (독립 다중 소스 파일): **PASS**
   - `TestCompilerModuleProjectBuild` (`go.mod` 패키지 모듈): **PASS**
3. **바이너리 빌드**:
   - macOS: `bin/tg` (성공)
   - Windows: `bin/tg.exe` (크로스 컴파일 성공)

---

## 🚀 실행 가이드

```bash
# Turbo Go 디렉터리로 이동
cd /Users/maro/Projects/tg

# 1. 기본 실행
./bin/tg

# 2. 예제 코드 열기
./bin/tg examples/hello.go

# 3. Windows 환경 실행 (Windows 터미널 / PowerShell)
.\bin\tg.exe examples\hello.go
```
