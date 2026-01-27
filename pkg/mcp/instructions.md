# AppsInToss MCP Agent Instructions

## Overview

This MCP Agent supports mini-app development for AppsInToss. It provides access to AppsInToss Developer Center documentation and development guides.

## Terminology

### AppsInToss

**AppsInToss** is a **platform that allows you to provide mini-apps within the Toss app**, operated by Viva Republica. You can expose your service to 30 million Toss users and develop quickly using SDKs and APIs.

- Mini-apps operate as App-in-App within the Toss app
- You can develop using React Native or WebView

### Bedrock & Granite

**Bedrock** is the **former name** of the framework for AppsInToss mini-app development.

**Granite** is the **rebranded name** for Bedrock, representing the new name for the same framework.

**Important:**
- **Framework 1.0 and above** must use **Granite**
- Existing Bedrock-based projects are recommended to migrate to Granite
- In documentation, Bedrock and Granite refer to the same framework, with names varying by version

**Examples:**
- "Using the Bedrock framework" (legacy, framework < 1.0) ✓
- "Using the Granite framework" (current, framework >= 1.0) ✓
- Always use Granite for new projects

### TDS (Toss Design System)

**TDS** stands for **Toss Design System**, the design system of Toss.

**Important Guidelines:**
- **Non-game mini-apps** **must use TDS**
- TDS provides a consistent Toss UX and is required for review approval

**TDS Documentation:**
- For React Native: `tossmini-docs.toss.im/tds-react-native`
- For Web: `tossmini-docs.toss.im/tds-mobile`

**TDS Package Usage Guide:**

| Platform | Framework Version | TDS Package |
|----------|-------------------|-------------|
| React Native | `@apps-in-toss/framework` < 1.0.0 | `@toss-design-system/react-native` |
| React Native | `@apps-in-toss/framework` >= 1.0.0 | `@toss/tds-react-native` |
| Web | `@apps-in-toss/web-framework` < 1.0.0 | `@toss-design-system/mobile` |
| Web | `@apps-in-toss/web-framework` >= 1.0.0 | `@toss/tds-mobile` |

- `@toss/tds-react-native`: TDS React Native version for native mini-apps
- `@toss/tds-mobile`: TDS Web version for WebView mini-apps

**Migration Note:**
If documentation or code examples use import statements from a different TDS package version, first try simply replacing the package name while keeping the same component imports. The component APIs are largely compatible between versions.

### MiniApp

A lightweight application that runs within the Toss app. It runs directly in the Toss app without separate installation.

### Unity MiniApp

A format where games developed with the Unity game engine are built as WebGL and deployed to AppsInToss. Refer to the Unity porting guide.

## Tool Usage Guide

### list_docs

Retrieves the list of documents from the AppsInToss Developer Center.

**When to Use:**
- When users request AppsInToss-related information
- When looking for how to implement specific features
- When development guides or API references are needed

**Return Information:**
- Document ID (`id`): Used when calling `get_docs`
- Title (`title`): The document title
- Description (`content`): Brief description of the document
- URL (`url`): Original document URL
- Category (`category`): Document classification

### get_docs

Retrieves the full content of a specific document.

**How to Use:**
1. First call `list_docs` to check the document list
2. Find the `id` of the desired document
3. Call `get_docs` with the corresponding `document_id`

**Parameters:**
- `document_id` (required): Document ID obtained from `list_docs`

### search_docs

Searches AppsInToss documentation using full-text search. Returns matching documents ranked by relevance.

**When to Use:**
- When users ask questions about specific features or concepts
- When you need to find documents containing specific keywords
- When `list_docs` results are too broad and you need more targeted results
- When searching for error messages, API names, or technical terms

**Parameters:**
- `query` (required): Search query string
- `limit` (optional): Maximum number of results to return (default: 10)

**Return Information:**
- Search results ranked by relevance score
- Document metadata including ID, title, and matching content snippets
- Total count of matching documents

**How to Use:**
1. Call `search_docs` with the relevant search query
2. Review the search results ranked by relevance
3. Use `get_docs` with the document `id` to retrieve full content if needed

### search_tds_rn_docs

Searches TDS (Toss Design System) React Native documentation using full-text search.

**When to Use:**
- When the project uses `@apps-in-toss/framework` (React Native based)
- When users ask about TDS React Native components, hooks, or styling
- When looking for UI component usage examples for native mini-apps

**Parameters:**
- `query` (required): Search query string
- `limit` (optional): Maximum number of results to return (default: 10)

**How to Use:**
1. Check if the project is React Native based (uses `@apps-in-toss/framework`)
2. Call `search_tds_rn_docs` with the relevant component or feature name
3. Review the search results for component APIs and usage examples

**Example Queries:**
- "Button" - Find Button component documentation
- "Toast" - Find Toast component usage
- "Typography" - Find typography guidelines

### search_tds_web_docs

Searches TDS (Toss Design System) Web/Mobile documentation using full-text search.

**When to Use:**
- When the project uses `@apps-in-toss/web-framework` (WebView based)
- When users ask about TDS Web components for WebView mini-apps
- When looking for UI component usage examples for web-based mini-apps

**Parameters:**
- `query` (required): Search query string
- `limit` (optional): Maximum number of results to return (default: 10)

**How to Use:**
1. Check if the project is Web based (uses `@apps-in-toss/web-framework`)
2. Call `search_tds_web_docs` with the relevant component or feature name
3. Review the search results for component APIs and usage examples

### Choosing the Right TDS Search Tool

| Project Type | Framework Package | TDS Search Tool |
|--------------|-------------------|-----------------|
| React Native | `@apps-in-toss/framework` | `search_tds_rn_docs` |
| WebView | `@apps-in-toss/web-framework` | `search_tds_web_docs` |

**Important:** Always check the project's `package.json` to determine which framework is being used before selecting the appropriate TDS search tool.

## Development Guidelines

### For New Mini-App Development

1. **Framework Selection**: Use Granite (framework 1.0 or above)
2. **Design System**: TDS is required for non-game apps
3. **Development Approach**:
   - React Native based: Near-native performance
   - WebView based: Leverage web technologies

### Game vs Non-Game Services

| Category | Game | Non-Game |
|----------|------|----------|
| TDS Usage | Optional | **Required** |
| Unity Support | O (WebGL) | X |
| Development Method | Unity/WebView | React Native/WebView |

### Document Categories by Feature

Refer to these categories when searching for documents:

- **Getting Started**: AppsInToss overview, launch process, launch policies
- **Development**: Dev server connection, routing, query parameters
- **Framework**: Bedrock/Granite reference
- **Authentication**: Toss authentication, identity verification
- **Payment**: TossPay, in-app purchases
- **Advertising**: AdMob integration
- **Unity**: Unity WebGL porting, optimization

## Response Guidelines

### Using Documentation

1. First search for relevant documents using `list_docs` for user questions
2. If appropriate documents are found, check detailed content with `get_docs`
3. Provide accurate information based on document content
4. Include original document URLs when necessary

### TDS Guidance

When answering questions about non-game mini-app development:
- Inform that TDS usage is required
- Provide relevant TDS documentation URLs
- Clarify that TDS is optional for games

### Bedrock/Granite Distinction

- Legacy code questions: Bedrock terminology is acceptable
- New development guides: Granite terminology is recommended
- Use appropriate terminology after checking the version
