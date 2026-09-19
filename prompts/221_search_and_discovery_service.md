# 221 - Search & Discovery Service (OpenSearch / PostgreSQL / Catalog Engine)

## Purpose
The Search & Discovery Service empowers investors to explore, discover, filter, and analyze Indian equities, sovereign green bonds, index baskets, and tokenized security offerings across Growww's multi-platform client applications. In a fractional investment ecosystem, discoverability must extend beyond basic stock ticker lookup to encompass real-time pricing, company fundamentals, ESG ratings, sector themes (e.g. Nifty 50, EV Ecosystem, Renewable Energy), corporate action history, and on-chain Proof-of-Reserve backing status.

This service delivers sub-10ms auto-complete, typo-tolerant fuzzy search, and thematic filtering, ensuring an intuitive, highly responsive discovery experience while strictly adhering to SEBI advertising codes by maintaining neutral, unbiased ranking algorithms without promotional steering.

## What You Are Building
A high-performance Go microservice (`services/search-service`) backed by OpenSearch / Elasticsearch delivering:
- **Instantaneous Fuzzy Search & Auto-Complete:** Full-text search over company names, NSE/BSE stock symbols, ISINs, and sector classifications.
- **Thematic Screener & Filter Engine:** Multi-attribute filtering on Market Cap, P/E Ratio, Dividend Yield, ESG Score, and 1-Year Returns.
- **Catalog Synchronization Worker:** Real-time ingestion and indexing of corporate fundamentals, daily market caps, and newly listed security tokens.
- **On-Chain Reserve Attestation Indexing:** Indexes on-chain smart contract addresses and real-time physical share backing status for search results.
- **Artifacts Delivered:**
 - `services/search-service/cmd/server/main.go` - Go microservice entry point.
 - `services/search-service/internal/indexer/opensearch.go` - OpenSearch indexing and query builder.
 - `services/search-service/internal/syncer/catalog_syncer.go` - Database & Kafka catalog syncer.
 - `services/search-service/config/opensearch_mappings.json` - Custom OpenSearch analyzer and token mapping definitions.
 - `proto/growww/search/v1/search.proto` - Internal gRPC search contracts.

## Scope Boundaries
- **In Scope:**
 - Full-text search, prefix matching, n-gram auto-complete, and phonetic matching (Soundex/Metaphone).
 - Faceted search and multi-parameter filtering across fundamental financial indicators.
 - OpenSearch index lifecycle management and near-real-time index updates from Kafka.
 - Indexing on-chain contract addresses and proof-of-reserve verification timestamps.
- **Out of Scope / Handled Elsewhere:**
 - Live streaming WebSocket order book quotes (handled by Prompt 207).
 - User watchlist persistence (handled by Portfolio Service, Prompt 209).
 - Algorithmic trade execution (handled by Prompt 205).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing `opensearch-go/v2` client, `pgx/v5`, and `confluent-kafka-go`.
- **Justification:** Go provides optimal memory efficiency and sub-millisecond serialization overhead for processing high-frequency search and auto-complete queries from millions of active client devices, while OpenSearch provides battle-tested distributed inverted indexing, fuzzy matching, and multi-facet aggregations.
- **Dependencies & Libraries:**
 - OpenSearch 2.12+ (or Elasticsearch 8+) cluster with custom edge n-gram tokenizers.
 - PostgreSQL 16+ as the source of truth for company master records and financial fundamentals.
 - Redis 7.2+ for caching top trending tickers and frequent query results.
 - Apache Kafka 3.7+ for consuming catalog updates and price snapshots.

## Backend / Infra Touchpoints
- **OpenSearch 2.12+:** Primary search index (`securities_catalog_v1`).
- **PostgreSQL 16:** Relational security master tables (`security_master`, `company_fundamentals`).
- **Redis 7.2:** Hot search query cache (60-second TTL) and popular search term rankings.
- **Kafka Topics:**
 - Subscribes: `market.security.updated`, `corporate_action.announced`, `custody.holdings.snapshot`.
- **Hyperledger Besu (QBFT):** Reads verified token contract deployment addresses and on-chain Proof-of-Reserve registry records.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium network running QBFT consensus.
- **On-Chain Attribute Indexing:** The Search Service indexes on-chain metadata for each security:
 - ERC-3643 `DigitalSecurityToken.sol` smart contract address.
 - Current total token supply on Besu vs physical demat shares in NSDL/CDSL custody.
 - Latest Proof-of-Reserve attestation block number and verification status.
- **Transparent Discovery:** Users searching for any security (e.g. "Reliance Industries" or "TCS") can view its verified on-chain contract address and cryptographic proof-of-reserve badge directly in the search results card.
- **Zero PII:** Catalog search indexes contain strictly public financial and corporate asset information.

## Step-by-Step Build Instructions
1. Scaffold Go project under `services/search-service` with standardized directory layout and config management.
2. Define Protobuf definitions in `proto/growww/search/v1/search.proto` and generate Go gRPC stubs.
3. Configure OpenSearch index mapping with custom edge n-gram tokenizers for prefix auto-complete and fuzzy matching.
4. Set up PostgreSQL schema migrations for `security_master`, `company_profiles`, and `thematic_tags`.
5. Implement the bulk indexing pipeline loading historical securities data from PostgreSQL into OpenSearch.
6. Build the Auto-Complete query engine using edge n-gram matchers delivering results in $<5\text{ms}$.
7. Build the Thematic Filter engine supporting faceted filtering (Market Cap, Sector, ESG score, Dividend yield).
8. Implement the Kafka event consumer updating OpenSearch documents in near-real-time upon price or corporate action changes.
9. Integrate Redis query result caching for hot search terms to minimize OpenSearch cluster CPU load.
10. Integrate on-chain contract addresses and Proof-of-Reserve status metadata into the indexed document schema.
11. Add Prometheus metrics (`search_queries_total`, `search_latency_ms`, `opensearch_indexing_errors`) and health checks.
12. Write comprehensive unit and integration tests verifying typo tolerance, fuzzy ranking, facet accuracy, and cache invalidation.

## Interfaces / Contracts

### Protobuf Definition (`search.proto`)
```protobuf
syntax = "proto3";

package growww.search.v1;

option go_package = "github.com/growww/services/search-service/gen/v1;searchv1";

service SearchDiscoveryService {
  rpc AutoComplete (AutoCompleteRequest) returns (AutoCompleteResponse);
  rpc SearchSecurities (SearchSecuritiesRequest) returns (SearchSecuritiesResponse);
  rpc FilterSecurities (FilterSecuritiesRequest) returns (FilterSecuritiesResponse);
  rpc GetSecurityDetails (SecurityDetailsRequest) returns (SecurityDetailsResponse);
}

message AutoCompleteRequest {
  string query = 1;
  int32 limit = 2; // Default 10
}

message AutoCompleteResponse {
  repeated SearchSuggestion suggestions = 1;
  int64 latency_ms = 2;
}

message SearchSuggestion {
  string isin = 1;
  string symbol = 2;
  string company_name = 3;
  string sector = 4;
  string latest_price_inr = 5;
  string price_change_24h_percent = 6;
  string on_chain_contract_address = 7;
  bool is_reserve_verified = 8;
}

message SearchSecuritiesRequest {
  string query = 1;
  int32 page = 2;
  int32 page_size = 3;
  string sort_by = 4; // RELEVANCE / MARKET_CAP / RETURNS / POPULARITY
  string sort_order = 5; // ASC / DESC
}

message SearchSecuritiesResponse {
  repeated SecuritySummary results = 1;
  int64 total_results = 2;
  int32 total_pages = 3;
}

message FilterSecuritiesRequest {
  repeated string sectors = 1;
  string min_market_cap = 2;
  string max_market_cap = 3;
  string min_dividend_yield = 4;
  string min_esg_score = 5;
  int32 page = 6;
  int32 page_size = 7;
}

message FilterSecuritiesResponse {
  repeated SecuritySummary results = 1;
  int64 total_matched = 2;
  map<string, int64> sector_facet_counts = 3;
}

message SecuritySummary {
  string isin = 1;
  string symbol = 2;
  string company_name = 3;
  string sector = 4;
  string market_cap_inr = 5;
  string pe_ratio = 6;
  string dividend_yield = 7;
  string current_price_inr = 8;
  string on_chain_token_address = 9;
}

message SecurityDetailsRequest {
  string isin = 1;
}

message SecurityDetailsResponse {
  string isin = 1;
  string symbol = 2;
  string company_name = 3;
  string description = 4;
  string industry = 5;
  string market_cap_inr = 6;
  string eps = 7;
  string pe_ratio = 8;
  string pb_ratio = 9;
  string on_chain_token_address = 10;
  string proof_of_reserve_merkle_root = 11;
  uint64 proof_of_reserve_block = 12;
  int64 last_attested_at = 13;
}
```

### OpenSearch Index Mapping (`securities_catalog_v1.json`)
```json
{
  "settings": {
    "analysis": {
      "analyzer": {
        "autocomplete_analyzer": {
          "type": "custom",
          "tokenizer": "edge_ngram_tokenizer",
          "filter": ["lowercase", "asciifolding"]
        }
      },
      "tokenizer": {
        "edge_ngram_tokenizer": {
          "type": "edge_ngram",
          "min_gram": 2,
          "max_gram": 15,
          "token_chars": ["letter", "digit"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "isin": { "type": "keyword" },
      "symbol": { "type": "keyword", "fields": { "suggest": { "type": "text", "analyzer": "autocomplete_analyzer" } } },
      "company_name": { "type": "text", "analyzer": "autocomplete_analyzer", "fields": { "raw": { "type": "keyword" } } },
      "sector": { "type": "keyword" },
      "market_cap_inr": { "type": "scaled_float", "scaling_factor": 100 },
      "pe_ratio": { "type": "float" },
      "dividend_yield": { "type": "float" },
      "esg_score": { "type": "short" },
      "on_chain_token_address": { "type": "keyword" },
      "is_reserve_verified": { "type": "boolean" },
      "updated_at": { "type": "date" }
    }
  }
}
```

## Security & Compliance Notes
- **SEBI Advertising Code Compliance:** Search algorithms must be strictly neutral and merit-based. Promotional boosting, paid ranking adjustments, or misleading "top gainer" steering are prohibited by system policy.
- **SQL / Injection Defense:** All user search inputs are strictly sanitized and parametrized against query-string injection.
- **Read Replica Isolation:** Search queries are served exclusively from OpenSearch and PostgreSQL read-replicas, preventing heavy user discovery traffic from impacting transactional order execution engines.

## Acceptance Criteria
- [ ] Go search service compiles cleanly and connects to OpenSearch cluster.
- [ ] Auto-complete endpoint returns relevant security suggestions in $<5\text{ms}$ for 2-character prefixes.
- [ ] Typo-tolerant search successfully matches company names with up to 2 edit distances (Levenshtein).
- [ ] Faceted filtering correctly returns accurate facet counts and filtered security lists.
- [ ] On-chain token contract addresses and Proof-of-Reserve flags are accurately indexed and returned.
- [ ] Kafka catalog synchronization updates OpenSearch index documents in under 500ms from event receipt.
- [ ] Automated test suite achieves >=85% code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design), Prompt 207 (Market Data), Prompt 213 (Custody Adapter).
- **Subsequent / Parallel Tasks:** Prompt 512 (Flutter App Search UI), Prompt 601 (Web Client Search).
