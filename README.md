# G4S CRM

Vue 3 / TypeScript frontend in `crm-dashboard/`, Go / Gin API in `crm-api/`, and PostgreSQL database migrations.

**Start with [STATUS.md](STATUS.md)** for current readiness, priorities, task states and the next implementation batch.

- [Implementation review](docs/CRM_REVIEW.md): verified findings, module coverage, business logic and table/schema recommendations.
- [Review evidence](docs/review-evidence/README.md): checks, logs and reproduction notes.
- [Original project documentation](PROJECT_DOCUMENTATION.md): historical design context.
- [Original backend development plan](crm-dashboard/BACKEND_DEVELOPMENT.md): historical intent and API planning.

The original documents contain outdated statements and describe some features that are not implemented. The review and status file distinguish actual implementation from intended behavior.

The project is not production-ready. Follow the verified setup work in STATUS.md before using the old Makefile deployment instructions; migration tooling and first-admin provisioning are incomplete.
