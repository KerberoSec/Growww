# ADR-0028: In-Band WebSocket Re-Authentication & Zero-Disconnection Token Refresh

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** API Gateway Lead, Frontend Architect, Security Lead  

---

## 1. Context
Short-lived JWT authentication tokens (15-minute validity) expire during continuous 24/7 trading sessions. Forcibly closing WebSocket connections upon token expiry interrupts real-time Level-2 order book streaming, clears local DOM state, and exposes active traders to visual blind spots.

---

## 2. Decision
Implement an in-band WebSocket re-authentication protocol:
1. **Client Token Refresh:** 60 seconds prior to JWT expiration, the client app requests a new JWT token from `user-service` via REST/gRPC.
2. **In-Band Refresh Message:** The client sends an in-band message over the existing WebSocket connection:
   ```json
   {
     "action": "auth_refresh",
     "token": "<new_jwt_token>"
   }
   ```
3. **Gateway Verification:** The API Gateway validates the signature and claims of `<new_jwt_token>`, updates the connection context's expiration deadline, and responds with `{"status": "AUTH_SUCCESS"}` without disconnecting the socket or altering channel subscriptions.
4. **Grace Period:** If no valid refreshed token is received within 30 seconds of expiry, the gateway gracefully terminates the connection with WebSocket closure code 4401 (`TOKEN_EXPIRED`).

---

## 3. Consequences
- **Positive:** Enables uninterrupted 24/7 continuous market data and execution streaming for all mobile and web clients.
- **Trade-offs:** API Gateway connection state must maintain mutable token expiry timestamps.
