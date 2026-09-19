# Frontend & Client Application Architecture Specification

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Frontend & Mobile Engineering Architecture Group  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Application Topology & Security Boundaries

```
+----------------------------------------------------------------------------------------------------+
|                                    FRONTEND CLIENT APPLICATION MAP                                 |
|                                                                                                    |
|  +-----------------------+  +-----------------------+  +-------------------+  +-----------------+  |
|  |     growww_web        |  |     growww_admin      |  | growww_marketing  |  | growww_flutter  |  |
|  | (Retail & Inst. Web)  |  | (Back-Office Portal)  |  | (Landing & Blog)  |  | (iOS/Android/OS)|  |
|  |                       |  |                       |  |                   |  |                 |  |
|  | * Next.js 14 App Router| * Next.js 14 App Router|  | * Next.js Static  |  | * Flutter 3.22  |  |
|  | * `robots: noindex`   |  | * `robots: noindex`   |  | * `robots: index` |  | * Riverpod State|  |
|  | * Session / JWT Auth  |  | * WebAuthn FIDO2 MFA  |  | * SEO & OG Meta   |  | * growww_ui Lib |  |
|  | * Idempotent Submits  |  | * Four-Eyes Approvals |  | * Risk Disclosures|  | * Offline Cache |  |
|  +-----------------------+  +-----------------------+  +-------------------+  +-----------------+  |
+----------------------------------------------------------------------------------------------------+
```

---

## 2. Security & Authentication Architecture

### 2.1 Admin Portal Authentication Gate (W-02)
The back-office admin portal (`growww_admin`) implements a strict deny-by-default Edge Middleware gate:

```typescript
// apps/growww_admin/src/middleware.ts
import { NextResponse, type NextRequest } from 'next/server';

export async function middleware(req: NextRequest) {
  const token = req.cookies.get('growww_admin_session')?.value;
  if (!token) {
    return NextResponse.redirect(new URL('/login', req.url));
  }
  
  const mfaVerified = req.cookies.get('growww_admin_mfa')?.value;
  if (mfaVerified !== 'fido2_verified') {
    return NextResponse.redirect(new URL('/mfa', req.url));
  }
  
  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!login|mfa|_next/static|_next/image|favicon.ico).*)'],
};
```

Server-side operations additionally enforce cryptographic four-eyes verification ensuring the initiator and approver are distinct corporate identities.

---

## 3. SEO, Metadata & Statutory Disclosures

### 3.1 Marketing Disclosures & Regulatory Compliance (W-01)
The marketing landing page (`growww_marketing`) strictly limits advertised asset classes to tokenized fractional Indian equities and MCX commodities backed by licensed depository participants (NSDL/CDSL).

Every public page enforces statutory disclosure footers:
- SEBI Registration Numbers (Stock Broker, Depository Participant).
- Standard Regulatory Warning: *"Investments in securities market are subject to market risks, read all the related documents carefully before investing."*
- Direct link to SEBI SCORES / ODR grievance redressal portals.

### 3.2 Metadata & Robots Indexing Strategy (W-03)
- **Public Sites (`growww_marketing`):** Full OpenGraph, Twitter Cards, canonical tags, `sitemap.xml`, and `robots: { index: true, follow: true }`.
- **Trading & Admin Applications (`growww_web`, `growww_admin`):** Strictly configured with `robots: { index: false, follow: false }` to prevent indexing of authenticated and transactional surfaces.

---

## 4. UI Error Handling & Ambiguous State Resolution (W-04)

Trading workflows implement dedicated error boundaries resolving network timeout ambiguities:

```tsx
export function OrderErrorPanel({ orderRef }: { orderRef: string }) {
  return (
    <div className="rounded-lg border border-amber-500/20 bg-amber-950/10 p-6 text-amber-200">
      <h3 className="text-lg font-semibold">Order Status Pending Confirmation</h3>
      <p className="mt-2 text-sm text-amber-300/80">
        Network confirmation for reference <code className="font-mono">{orderRef}</code> was interrupted.
        Your order may have been accepted. Please inspect your active order book before retrying.
      </p>
      <div className="mt-4 flex gap-3">
        <a href="/orders" className="rounded bg-amber-600 px-4 py-2 text-xs font-bold text-white">
          Inspect Order Book
        </a>
      </div>
    </div>
  );
}
```

Combined with client-side `Idempotency-Key` tracking (D-03), subsequent user retries are guaranteed to prevent duplicate order placement.

---

## 5. Regional Formatting, Accessibility & Performance (W-05 to W-12)

1. **Indian Currency Formatting (W-12):** All price and valuation strings use `Number.toLocaleString('en-IN')` rendering standard Lakh and Crore separators:
   $$\text{₹ 12,34,567.89}$$
2. **Accessibility Standards (W-08):** WCAG 2.1 AA compliance with skip navigation links, `focus-visible` ring outlines, high-contrast ratios (minimum 4.5:1), and CSS `prefers-reduced-motion` compliance.
3. **Data Protection Consent Management (W-09):** DPDP Act 2023 compliant consent banner capturing explicit opt-ins before firing non-essential analytics cookies.
4. **Performance & Bundle Budgets (W-11):** Next.js initial route bundles are budgeted at <150 KB gzip, verified in continuous integration.
