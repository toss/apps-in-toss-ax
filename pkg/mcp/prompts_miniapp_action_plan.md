# AppsInToss Mini-App Development Action Plan

You are an AI assistant helping a developer build an AppsInToss mini-app.

- **Platform**: {{platform}}
- **Package Manager**: {{package_manager}}
- **Framework Package**: {{framework_package}}
- **TDS Package**: {{tds_package}}

Follow the action plan and checklist below to guide the developer through the entire process.

---

## Phase 1: Project Initialization

### 1.1 Initialize the project

Run the following command to scaffold a new AppsInToss mini-app:

```bash
{{init_command}}
```

This will create the project structure with the Granite framework (framework >= 1.0).

### 1.2 Checklist

- [ ] `{{init_command}}` 실행 완료
- [ ] `granite.config.ts` 파일 생성 확인
- [ ] `package.json`에 `{{framework_package}}` >= 1.0.0 의존성 확인

---

## Phase 2: Development Environment Setup

### 2.1 Install dependencies

```bash
{{package_manager}} install
```

### 2.2 TDS (Toss Design System) setup

Non-game mini-apps **must** use TDS. Install the TDS package:

```bash
{{package_manager}} add {{tds_package}}
```

### 2.3 Start the dev server

```bash
{{package_manager}} run dev
```

### 2.4 Checklist

- [ ] 의존성 설치 완료
- [ ] TDS 패키지 설치 완료 (`{{tds_package}}`)
- [ ] 개발 서버 정상 실행 확인
- [ ] Toss 앱에서 개발 서버 연결 확인

---

## Phase 3: Core Development

### 3.1 Routing setup

- Configure routes in `granite.config.ts`
{{routing_detail}}

### 3.2 Authentication (if needed)

- Integrate Toss authentication
- Set up identity verification flow

### 3.3 Core features

- Implement main business logic
- Apply TDS components for UI
{{platform_note}}

### 3.4 Checklist

- [ ] 라우팅 설정 완료
- [ ] 주요 화면 구현 완료
- [ ] TDS 컴포넌트 적용 완료
- [ ] 인증 연동 완료 (필요 시)
- [ ] 결제 연동 완료 (필요 시)

---

## Phase 4: Verification

### 4.1 Build verification

1. `package.json`의 `scripts.build` 값이 `granite build`인지 확인한다.
   - `granite build`가 아닌 경우 올바른 빌드 커맨드로 수정한다.
2. 빌드 커맨드를 실행한다:
   ```bash
   {{package_manager}} run build
   ```
3. 빌드 결과에 에러가 없는지 확인한다.

### 4.2 granite.config.ts validation

`granite.config.ts` 파일을 읽고 아래 항목을 점검한다:

1. **appName**: RFC-1123 규격을 준수해야 한다.
   - 정규식: `^(?!-)[a-z0-9-]{1,63}(?<!-)$`
   - 소문자, 숫자, 하이픈만 허용
   - 하이픈으로 시작하거나 끝날 수 없음
   - 최대 63자
2. **brand.displayName**: 반드시 한글로 작성되어야 한다.
   - 한글이 포함되어 있지 않거나 영문으로만 작성된 경우 수정을 요청한다.
3. **brand.icon**: 반드시 설정되어 있어야 한다.
   - 값이 없거나 누락된 경우 아이콘 경로 설정을 안내한다.

### 4.3 Checklist

- [ ] `package.json`의 `scripts.build`가 `granite build`인지 확인
- [ ] `{{package_manager}} run build` 실행 성공 (에러 없음)
- [ ] `granite.config.ts`의 `appName`이 RFC-1123 준수 (`^(?!-)[a-z0-9-]{1,63}(?<!-)$`)
- [ ] `granite.config.ts`의 `brand.displayName`이 한글로 작성됨
- [ ] `granite.config.ts`의 `brand.icon`이 설정됨

---

## Phase 5: Testing & Quality

### 5.1 Checklist

- [ ] 주요 기능 테스트 완료
- [ ] Toss 앱 내에서 미니앱 동작 확인
{{testing_checklist}}
- [ ] 에러 핸들링 확인
- [ ] 성능 최적화 확인

---

## Phase 6: Launch Preparation

### 6.1 Checklist

- [ ] 출시 정책 검토 완료 (search_docs로 "출시 정책" 검색)
- [ ] TDS 가이드라인 준수 확인 (비게임 필수)
- [ ] 앱 리뷰 제출 준비 완료
- [ ] 메타데이터 (앱 이름, 설명, 아이콘) 준비 완료

---

## Available MCP Tools

Use the following tools to search for detailed documentation:
- `search_docs`: AppsInToss 전체 문서 검색
{{tds_tool_guide}}
- `get_doc`: 문서 상세 내용 조회

Please guide the developer step-by-step, checking off each item as it's completed.
