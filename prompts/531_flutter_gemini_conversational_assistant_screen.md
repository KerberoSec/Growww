# 515 - Flutter Gemini Conversational Assistant Screen & Interactive Drawer (Riverpod / Markdown / WebRTC)

## Purpose
Navigating multi-asset tokenized capital markets, fractional share ownership, blockchain-notarized depository reserves, and complex regulatory documentation can be daunting for retail and institutional investors. Investors require an intuitive, instant, on-demand conversational intelligence companion accessible across all platform screens. Whether reviewing a specific security on the detail screen, exploring portfolio diversification, verifying Proof-of-Reserve backing on Hyperledger Besu, or reviewing transaction fee deductions, users need immediate, plain-language financial clarity.

The **Flutter Gemini Conversational Assistant Screen & Interactive Drawer** (`apps/growww_flutter/lib/features/gemini_assistant/`) provides a multi-platform, multimodal AI interface embedded within the Growww application. It functions both as a slide-over / bottom sheet interactive drawer accessible from any trading view and as an expansive full-screen educational companion. The assistant renders token-by-token streaming markdown with embedded financial widgets, interactive scenario sliders for hypothetical portfolio stress testing, animated chain-of-custody visual flowcharts, voice input via WebRTC and speech-to-text, one-tap prompt chips, and complete fee transparency explaining the platform's flat 0.00% transaction fee (No fee at all) (0.00% fee at launch (governed by FeeController.sol) split).

## What You Are Building
A responsive, high-performance Flutter feature module (`apps/growww_flutter/lib/features/gemini_assistant/`) supporting Android, iOS, macOS, Windows, Linux, and Web:
- **Dual Presentation Modes:**
  - *Contextual Modal Drawer / Slide-Over Sheet:* Floating overlay that slides over active screens (such as Security Detail or Order Placement) to explain metrics without disrupting user workflow.
  - *Full-Screen Conversational Companion:* Dedicated primary tab with rich chat history, document summarization, voice session modes, and scenario exploration.
- **Streaming Markdown & Equation Renderer:** Hardware-accelerated markdown rendering pipeline supporting formatted tables, callout banners, clickable citation badges, and mathematical formulas via KaTeX.
- **Embedded Interactive Scenario Sliders:** Dynamic visual widgets embedded directly within chat bubbles that allow users to drag sliders (such as interest rate changes of +/- 200 bps or index shocks of +/- 15%) and visualize hypothetical portfolio impacts in real time.
- **Chain-of-Custody Visual Flowcharts:** Interactive, animated step diagram rendering the asset provenance lifecycle: `Investor` $\rightarrow$ `Digital Token` $\rightarrow$ `Hyperledger Besu Ledger` $\rightarrow$ `NSDL / CDSL Demat Custody`.
- **Multimodal Voice Input & Waveform Visualizer:** Audio capture pipeline with real-time WebRTC / WebSocket streaming, animated voice frequency visualizer, and seamless text-to-speech audio playback.
- **Context-Aware One-Tap Prompt Chips:** Horizontally scrolling quick-action chips that dynamically adapt based on the user's active screen (such as "Explain what I am investing in", "Where is the physical share held?", "Verify Proof of Reserve", "Why was 0.00% fee (No fee at all) charged?").
- **Transparent Fee Breakdown Card:** Specialized widget presenting sub-paise itemizations of the platform's 0.00% transaction fee (No fee at all) and proving the zero-AUM holding fee guarantee.
- **Riverpod State Management Layer:** Robust state machines handling streaming response buffers, voice sessions, message history persistence, and scenario state.

## Scope Boundaries
- **In Scope:**
  - Dual-mode UI presentation (Slide-over drawer and Full-Screen view) across mobile and desktop breakpoints.
  - Streaming markdown rendering with syntax highlighting, citation tags, and financial formulas.
  - Custom Flutter painters for chain-of-custody flowcharts and scenario shock sliders.
  - Real-time voice recording, WebSockets audio streaming, and animated audio waveform visualization.
  - Contextual prompt chip generation and deep-linking into app features.
  - Offline message caching and local encrypted conversation persistence.
  - Display of non-advisory compliance notices and statutory SEBI risk disclaimers.
- **Out of Scope / Handled Elsewhere:**
  - Backend Gemini LLM orchestration, RAG retrieval, and guardrail classification (handled in Prompt 252).
  - Executing actual buy/sell market orders (handled in Prompt 509).
  - Primary portfolio balance aggregation and calculations (handled in Prompt 209).
  - Direct blockchain node RPC interactions (handled in Prompt 252 / Prompt 308).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ utilizing `flutter_riverpod` (v2.5+) for reactive state management.
- **Markdown & Math Rendering:** `flutter_markdown: ^0.7.2` extended with custom element builders and `flutter_math_fork: ^0.7.2` for crisp LaTeX formula typography.
- **Vector & Micro-Animations:** `lottie: ^3.1.2` for smooth AI thinking indicator animations and `rive: ^0.13.0` for interactive custody flowcharts.
- **Voice & Audio Telemetry:** `record: ^5.1.2` for cross-platform low-latency microphone audio capture and `audioplayers: ^6.0.0` for AI voice playback.
- **Network Transport & SSE:** `dio: ^5.4.3+1` configured with `fetch` adapter and `web_socket_channel: ^3.0.0` for bi-directional voice and streaming text tokens.
- **Local Persistence & Security:** `isar: ^3.1.0` or `flutter_secure_storage: ^9.2.2` for caching encrypted conversation transcripts locally on device.

## Backend / Infra Touchpoints
- **Gemini Advisor Service (Prompt 252):** Connects to `POST /api/v1/advisor/chat/stream` via Server-Sent Events (SSE) for token deltas and widget triggers.
- **Voice Stream WebSocket (`wss://ws.growww.in/v1/advisor/voice`):** Streams raw PCM audio packets and receives synthesized audio chunks.
- **Proof-of-Reserve Registry API (Prompt 243 / Prompt 308):** Fetches verifiable Merkle roots and custody balances to hydrate custody verification dialogs.
- **Portfolio Service (Prompt 209):** Supplies asset allocation weights for the scenario shock simulator without disclosing investor PII to external APIs.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Chain-of-Custody Verification Overlay:** Tapping any on-chain citation badge or custody node opens a modal displaying the exact Hyperledger Besu block height, the Merkle root hash on `ProofOfReserveRegistry.sol`, and the depository settlement batch ID.
- **100% Asset Backing Assurance:** The UI explicitly communicates that all fractional digital tokens are 1:1 backed by underlying physical securities held with SEBI-registered depositories (NSDL/CDSL).
- **Zero On-Chain PII Guarantee:** The assistant UI reminds users that their personal identity, chat transcripts, and confidential credentials are never written to the blockchain.

## Assistant UI Architecture & Multimodal Interaction Flow

### 1. Presentation Modes & Adaptive Layout Strategy
The Gemini Assistant adapts gracefully to device form factors and user contexts:
- **Mobile Handsets (< 600dp width):**
  - Contextual Trigger: Opens as a draggable modal bottom sheet (`showModalBottomSheet`) with snap heights at 50% and 90% screen heights.
  - Dedicated View: Full-screen route with bottom navigation bar integration.
- **Desktop & Web (> 600dp width):**
  - Contextual Trigger: Smooth slide-out right drawer (420dp fixed width) with backdrop blur, allowing side-by-side interaction with trading charts.
  - Dedicated View: Expansive 2-column layout (left sidebar for conversation history and document uploads, right panel for interactive chat and scenario visualizations).

```
+-----------------------------------------------------------------------------+
|                          GROWWW TRADING PLATFORM                            |
+------------------------------------------------------+----------------------+
|  Security Detail: RELIANCE INDUSTRIES (₹2,940.50)    |  GEMINI ASSISTANT    |
|  [Chart] [Order Book] [Corporate Actions]            |  (Slide-Over Drawer) |
|                                                      |                      |
|  Current Holdings: 12.45 Shares                      |  "Where is my asset  |
|  Total Value: ₹36,609.22                             |   held?"             |
|                                                      |                      |
|  +------------------------------------------------+  |  [Chain of Custody]  |
|  | Tap [Ask Gemini] to explain 1:1 custody backing |  |  Investor            |
|  +------------------------------------------------+  |      |               |
|                                                      |      v               |
|                                                      |  Hyperledger Besu    |
|                                                      |      |               |
|                                                      |      v               |
|                                                      |  NSDL / CDSL Demat   |
|                                                      |  (100% Backed)       |
|                                                      |                      |
|                                                      |  [0.00% fee (No fee at all) Detail]  |
|                                                      |  [Voice Mic: Active] |
+------------------------------------------------------+----------------------+
```

### 2. Embedded Interactive Financial Widgets
When the backend streaming engine emits an `embedded_widget` payload, the chat rendering engine inflates dedicated Flutter widgets inline with text:
1. **Scenario Shock Slider (`ScenarioShockSliderWidget`):** Provides interactive slider handles for interest rate shift ($\Delta y$) and equity index shock ($\Delta I$). Dragging the slider dynamically re-computes simulated portfolio valuation shifts on the client using local state without requiring additional API round-trips.
2. **Chain-of-Custody Flowchart (`CustodyFlowchartWidget`):** An animated vertical or horizontal node graph illustrating the end-to-end custody link, featuring green checkmarks on verified Merkle proofs and tap-to-expand details for NSDL/CDSL accounts.
3. **Transparent Fee Breakdown Card (`FeeTransparencyCard`):** Visually renders the flat 0.00% transaction fee (No fee at all) with a pie/bar chart depicting the exact statutory and system split: Treasury reserve, Core Settlement Guarantee Fund (SGF), and Investor Protection Fund (IPF) per FeeController governance.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Feature Directory:** Create `apps/growww_flutter/lib/features/gemini_assistant/` with subdirectories `presentation/screens/`, `presentation/widgets/`, `presentation/controllers/`, `domain/models/`, `domain/interfaces/`, `data/repositories/`, and `data/datasources/`.
2. **Define Domain Enums & Models:** Author immutable models in `domain/models/`: `ChatMessage`, `MessageRole`, `CitationSource`, `ScenarioShockData`, `CustodyNodeData`, `PromptChipItem`, and `VoiceSessionState`.
3. **Implement Assistant Repository & SSE Parser:** Create `GeminiAssistantRepository` using `Dio` and Server-Sent Events to parse token deltas, inline citation chunks, and JSON widget payloads.
4. **Implement Assistant Controller (`GeminiAssistantNotifier`):** Build Riverpod `AutoDisposeAsyncNotifier` orchestrating message history, streaming buffer aggregation, active screen context injection, and error recovery.
5. **Build Streaming Markdown Renderer:** Create `StreamingMarkdownView` widget utilizing `flutter_markdown` with customized syntax highlighters, bullet styles, and KaTeX math builders for financial equations.
6. **Build Chain-of-Custody Flowchart Widget:** Implement `CustodyFlowchartWidget` using custom `CustomPainter` to draw interactive custody paths connecting Investor $\rightarrow$ Digital Token $\rightarrow$ Hyperledger Besu $\rightarrow$ Demat Custody.
7. **Build Scenario Shock Slider Widget:** Implement `ScenarioShockSliderWidget` with smooth logarithmic/linear sliders, updating real-time simulated P&L curves and risk labels.
8. **Build Transparent Fee Breakdown Card:** Implement `FeeTransparencyCard` displaying the exact 0.00% transaction fee (No fee at all) split (0.00% fee at launch; future fee parameters governed by FeeController.sol) with zero-AUM fee badges.
9. **Implement Dynamic Prompt Chips Bar:** Create `ContextualPromptChipsBar` rendering horizontally scrolling chips tailored to current screen context (e.g. "Explain Fractional Shares", "Verify Proof of Reserve").
10. **Build Voice Input & Waveform Visualizer:** Implement `VoiceInteractionWidget` integrating `record` audio capture, real-time PCM audio streaming over WebSockets, and pulsating waveform animations.
11. **Build Contextual Slide-Over Drawer Shell:** Create `GeminiDrawerSheet` supporting adaptive desktop slide-out drawers and mobile modal bottom sheets with smooth enter/exit transitions.
12. **Build Full-Screen Assistant Screen:** Create `GeminiAssistantFullScreen` with message thread list, document upload buttons for DRHP summarization, and search history filters.
13. **Implement Non-Advisory Guardrail Alert Banner:** Build contextual disclaimer card warning users that all insights are educational only and strictly non-advisory.
14. **Write Comprehensive Test Suite:** Author unit tests for streaming message assembly, widget golden tests for light/dark themes, and integration tests for drawer open/close transitions.

## Interfaces / Contracts

### Domain Data Models (`lib/features/gemini_assistant/domain/models/assistant_models.dart`)
```dart
enum MessageRole {
  user,
  assistant,
  system,
}

enum AssistantPresentationMode {
  bottomSheetDrawer,
  sideSlideOverDrawer,
  fullScreen,
}

enum EmbeddedWidgetType {
  custodyFlowchart,
  scenarioSlider,
  feeBreakdownCard,
  documentSummaryCard,
}

class CitationSource {
  final String title;
  final String url;
  final String documentType;
  final String? onChainTxHash;

  const CitationSource({
    required this.title,
    required this.url,
    required this.documentType,
    this.onChainTxHash,
  });
}

class EmbeddedWidgetData {
  final EmbeddedWidgetType widgetType;
  final Map<String, dynamic> rawParameters;

  const EmbeddedWidgetData({
    required this.widgetType,
    required this.rawParameters,
  });
}

class ChatMessage {
  final String messageId;
  final MessageRole role;
  final String content;
  final bool isStreaming;
  final bool isGuardrailWarning;
  final List<CitationSource> citations;
  final EmbeddedWidgetData? embeddedWidget;
  final DateTime timestamp;

  const ChatMessage({
    required this.messageId,
    required this.role,
    required this.content,
    this.isStreaming = false,
    this.isGuardrailWarning = false,
    this.citations = const [],
    this.embeddedWidget,
    required this.timestamp,
  });
}

class PromptChipItem {
  final String chipId;
  final String displayText;
  final String promptQuery;
  final String iconAsset;

  const PromptChipItem({
    required this.chipId,
    required this.displayText,
    required this.promptQuery,
    required this.iconAsset,
  });
}

class ScenarioShockParams {
  final double interestRateDeltaBps;
  final double equityMarketShockPct;
  final double commodityShockPct;

  const ScenarioShockParams({
    required this.interestRateDeltaBps,
    required this.equityMarketShockPct,
    required this.commodityShockPct,
  });
}
```

### Riverpod State Contract (`lib/features/gemini_assistant/presentation/controllers/assistant_state.dart`)
```dart
class GeminiAssistantState {
  final List<ChatMessage> messageHistory;
  final bool isGenerating;
  final bool isVoiceActive;
  final String activeScreenContext;
  final String? activeIsin;
  final List<PromptChipItem> contextualPromptChips;
  final double currentVoiceInputAmplitude;
  final String? activeErrorMessage;

  const GeminiAssistantState({
    this.messageHistory = const [],
    this.isGenerating = false,
    this.isVoiceActive = false,
    this.activeScreenContext = 'HOME_DASHBOARD',
    this.activeIsin,
    this.contextualPromptChips = const [],
    this.currentVoiceInputAmplitude = 0.0,
    this.activeErrorMessage,
  });
}

abstract class IGeminiAssistantRepository {
  Stream<ChatMessage> streamQuery({
    required String prompt,
    required String screenContext,
    String? isinContext,
  });

  Future<void> sendVoiceChunk(List<int> pcmBytes);
  Future<void> clearConversationHistory();
  Future<List<PromptChipItem>> fetchContextualChips(String screenContext);
}
```

## Security & Compliance Notes
- **SEBI Non-Advisory Statutory Disclaimer:** A persistent, non-dismissible banner is displayed at the top of every assistant conversation session: *"Gemini Assistant provides financial education and data intelligence only. It does not provide investment advice, buy/sell recommendations, or price targets. Real securities are held 1:1 in custody with SEBI-registered depositories."*
- **Local Transcript Encryption:** All cached chat histories saved to client storage are encrypted using AES-GCM-256 with encryption keys backed by the device's secure hardware enclave (Android Keystore / iOS Keychain).
- **Client-Side PII Redaction:** Before transmitting any user query or voice transcript to the assistant backend, client-side regex filters strip PAN cards, Aadhaar numbers, and bank account numbers.
- **Transparent Fee Model Display:** All fee explanations explicitly display the 0.00% (No fee at all) platform fee (0 bps (0.00% fee at launch) on turnover) with its 0.00% fee at launch (governed by FeeController.sol) split, confirming zero holding fees, zero AUM management fees, and zero hidden markups.

## Acceptance Criteria
- [ ] Assistant opens seamlessly as a slide-over drawer on desktop (> 600dp) and as a modal bottom sheet on mobile (< 600dp).
- [ ] Streaming markdown renders token deltas in real time without screen flicker or dropped frames (maintaining 60 FPS).
- [ ] Chain-of-custody visual flowchart renders interactive nodes (Investor $\rightarrow$ Digital Token $\rightarrow$ Hyperledger Besu $\rightarrow$ Demat Custody NSDL/CDSL) with working tap-to-verify modals.
- [ ] Scenario shock slider updates simulated portfolio values dynamically as the slider thumb is dragged.
- [ ] Voice interaction records microphone input, renders animated waveform fluctuations, and plays back synthesized audio.
- [ ] Contextual prompt chips automatically refresh when the active screen context changes (e.g. switching from Portfolio to Security Detail).
- [ ] Fee breakdown widget accurately itemizes the flat 0.00% transaction fee (No fee at all) and its 0.00% fee launch policy statutory allocation.
- [ ] All chat transcripts on device are stored in encrypted format.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero raw application code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `501` (Flutter Project Scaffolding), Prompt `502` (State Management Architecture), Prompt `503` (Design System & Theming), Prompt `521` (Local Secure Storage).
- **Backend Dependency:** Prompt `252` (Gemini Financial Intelligence & Literacy Service), Prompt `207` (Market Data Service), Prompt `209` (Portfolio Service).
- **Enables:** Interactive AI assistance across Prompts `506` (Home Dashboard), `508` (Security Detail Screen), `509` (Order Placement Flow), and `510` (Portfolio Holdings).
