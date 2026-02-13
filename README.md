# AX (AppsInToss eXperience)

AI 어시스턴트가 AppsInToss 미니앱 개발을 도울 수 있도록 설계된 MCP(Model Context Protocol) 서버입니다.

## 주요 기능

AX는 AI 어시스턴트에게 AppsInToss Developer Center의 문서와 예제를 제공하여, 토스 미니앱 개발에 대한 정확하고 컨텍스트 기반의 지원을 가능하게 합니다.

### MCP Tools

| 도구 | 설명 |
|------|------|
| `search_docs` | AppsInToss 문서 검색 |
| `get_doc` | 검색 결과의 문서 전체 내용 조회 |
| `search_tds_rn_docs` | TDS React Native 문서 검색 |
| `get_tds_rn_doc` | TDS React Native 문서 전체 내용 조회 |
| `search_tds_web_docs` | TDS Web 문서 검색 |
| `get_tds_web_doc` | TDS Web 문서 전체 내용 조회 |
| `list_examples` | 코드 예제 목록 조회 |
| `get_example` | 특정 예제 코드 조회 |

### 지원 문서

- **AppsInToss Developer Center** - 미니앱 개발 가이드
- **TDS React Native** - 토스 디자인 시스템 (React Native)
- **TDS Web** - 토스 디자인 시스템 (WebView)
- **코드 예제** - 실제 구현 예제 모음

## 설치

### Homebrew (macOS/Linux)

```bash
brew tap toss/tap
brew install ax
```

### Scoop (Windows)

```powershell
scoop bucket add toss https://github.com/toss/scoop-bucket
scoop install ax
```

### npm

```bash
npm install -g @apps-in-toss/ax
```

## 사용법

### MCP 서버 시작

```bash
ax mcp start
```

### Cursor/Claude에서 사용

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/ko/install-mcp?name=apps-in-toss&config=eyJjb21tYW5kIjoiYXggbWNwIHN0YXJ0In0%3D)

또는 `.cursor/mcp.json` / `claude_desktop_config.json`에 다음을 추가하세요:

```json
{
  "mcpServers": {
    "apps-in-toss": {
      "command": "ax",
      "args": [
        "mcp", "start"
      ]
    }
  }
}
```

## 프로젝트 구조

```
apps-in-toss-ax/
├── cmd/                    # CLI 명령어 정의
├── internal/               # 내부 패키지
│   ├── httputil/          # HTTP 유틸리티
│   └── utils/             # 공통 유틸리티
├── pkg/                    # 핵심 패키지
│   ├── app/               # 애플리케이션 진입점
│   ├── docs/              # 문서 관리
│   ├── docid/             # 문서 ID 생성
│   ├── fetcher/           # HTTP 클라이언트
│   ├── llms/              # llms.txt 파서
│   ├── mcp/               # MCP 서버 구현
│   └── search/            # 문서 검색 엔진
├── tools/                  # 배포 도구
│   ├── publish/           # 패키지 매니저 배포
│   └── shared/            # 공유 유틸리티
├── scripts/                # 스크립트
├── main.go                # 메인 진입점
├── go.mod                 # Go 모듈 정의
├── Makefile               # 빌드 명령어
└── package.json           # npm 패키지 정의
```

## 개발

### 빌드

```bash
make build
# 또는
go build -o ax
```

### 테스트

```bash
go test ./...
```

### 로컬 실행

```bash
./ax mcp
```

## AppsInToss란?

**AppsInToss**는 토스 앱 내에서 미니앱을 제공할 수 있는 플랫폼입니다. 3,000만 토스 사용자에게 서비스를 노출하고, SDK와 API를 활용해 빠르게 개발할 수 있습니다.

### 주요 개념

| 용어 | 설명 |
|------|------|
| **Granite** | 미니앱 개발 프레임워크 (구 Bedrock, v1.0+) |
| **TDS** | Toss Design System - 비게임 미니앱 필수 |

### 개발 방식

- **React Native**: 네이티브에 가까운 성능
- **WebView**: 웹 기술 활용

## 관련 링크

- [AppsInToss Developer Center](https://developers-apps-in-toss.toss.im)
- [TDS React Native](https://tossmini-docs.toss.im/tds-react-native)
- [TDS Web](https://tossmini-docs.toss.im/tds-mobile)
