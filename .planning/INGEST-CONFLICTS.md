## Conflict Detection Report

### BLOCKERS (0)

(none)

### WARNINGS (0)

(none)

### INFO (4)

[INFO] Single-doc ingest — no inter-document conflict surface
  Note: One classified doc consumed (HAND_OFF.md, type=SPEC). With a single source there is no LOCKED-vs-LOCKED, no competing PRD acceptance variants, and no SPEC-vs-ADR contradiction surface to evaluate. All conflict-detection passes (LOCKED-vs-LOCKED, PRD overlap, SPEC-vs-ADR, lower-vs-higher precedence) are trivially satisfied.
  source: /home/saghaulor/code/security_reviewer/.planning/intel/classifications/HAND-OFF-7f3e9c2a.json

[INFO] Manifest precedence override applied
  Note: Classification declared manifest_override=true with precedence=0 (highest). The default precedence ordering [ADR, SPEC, PRD, DOC] is honored, but HAND_OFF.md's effective precedence of 0 puts it ahead of any future ADR ingested in this set. No conflict because no other doc was ingested.
  source: /home/saghaulor/code/security_reviewer/.planning/intel/classifications/HAND-OFF-7f3e9c2a.json (fields: precedence, manifest_override)

[INFO] §4 Decisions log (D1–D18) lifted to intel/decisions.md as LOCKED
  Note: Per orchestrator instruction, the eighteen entries in HAND_OFF.md §4 are treated as locked design decisions for downstream roadmapping, even though the source document is SPEC-classified (not ADR) and its classification carries locked=false. The lock state was promoted by the ingest orchestrator, not by the classifier. Downstream consumers must treat decisions.md entries as non-overridable until the user explicitly unlocks one.
  source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (Decisions log)
  source: /home/saghaulor/code/security_reviewer/.planning/intel/decisions.md

[INFO] §5 Open questions (Q1–Q10) captured as context, not blockers
  Note: HAND_OFF.md §5 lists ten open questions (Q1–Q10) that the implementing session must resolve during build. Per orchestrator instruction these are integration details (model identifiers, MCP tool names, Semgrep license provisioning, OpenGrep build pinning, etc.) and are recorded in intel/context.md under "Open questions" rather than surfaced as workflow blockers. They do not gate roadmap routing; they will surface naturally in Phase 2–4 work items.
  source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §5
  source: /home/saghaulor/code/security_reviewer/.planning/intel/context.md ("Topic: Open questions")
