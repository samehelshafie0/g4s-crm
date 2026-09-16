# Browser observations — isolated CRM

Date: 2026-09-16. Origin: `http://127.0.0.1:5174`, proxying to the real review API on port 18080. Synthetic PostgreSQL fixtures only. Chrome was controlled through the browser tool.

1. **Login:** successful login with the synthetic admin. The login page renders the sidebar/header for the full app while signed out; this comes from App always rendering MainLayout.
2. **Dashboard:** API-backed KPI cards showed one open quote worth SAR 999 (the deliberately modified review fixture). Recent quotes, activity, expiring quotes and representative performance still showed embedded sample records and relative times. This confirms the hybrid data model in the source.
3. **Quotations, desktop:** six embedded sample quotes rendered instead of the single review-database quote. The dark visual system and status badges are consistent. Quote IDs wrapped over three lines; the wide financial table and many icon actions leave insufficient room for identity/customer text.
4. **Quotations, 390 × 844:** sidebar stayed expanded and most content was off-screen. Read-only DOM dimensions: `innerWidth = 390`, `document.documentElement.scrollWidth = 1686`. Viewport override was reset after the check.
5. **Customers:** opening the screen with a real customer having no sites/contacts produced an empty main area and Vue render exception: `TypeError: Cannot read properties of undefined (reading 'length')`. API omits empty association arrays; template assumes them present.
6. **Hard refresh:** `/auth/me` returned ID/email/role without first/last names. Header showed `undefined undefined`; console reported `TypeError: Cannot read properties of undefined (reading '0')` in `stores/auth.ts`, in addition to the customer-array error.

Screenshots were visually inspected during the session but are not retained in this evidence folder. This pass did not cover every modal, every role, all themes, screen-reader behavior, printing, or complete performance profiling. The source-based recommendations identify that remaining work explicitly.
