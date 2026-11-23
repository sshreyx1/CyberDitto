# Acceptable use

CyberDitto runs real attacker techniques. These terms apply to anyone operating an instance of it.

## You may

- Collect configuration from machines **you own or have written permission to assess**, and upload it to an instance you control.
- Build twins from that configuration and run the curated attack catalog **against those twins, inside an isolated lab network**.

## You may not

- Point CyberDitto, its collector, or its catalog at any system you are not authorised to test.
- Weaken or bypass the lab safety invariants (docs/MASTER_PLAN.md §6.4): twin-only targeting, lab isolation checks, the destructive-test denylist, approver authorisation, run limits and the kill switch, the pinned and verified catalog, and the rule that drafts (including AI-generated tests) never execute.
- Connect a lab network to a production network or the internet.
- Use pseudonymised exports to re-identify people, or upload snapshots that contain personal data you have no right to process.

## What the product does to enforce this

- Every attack run needs an approver to acknowledge, per run: *"I am authorised to run these attack tests against this lab twin, and I understand they execute real attacker techniques inside the isolated lab."* The acknowledgement, who gave it, and the preflight results are stored with the run and written to the hash-chained audit log.
- Runs can only target twin records; there is no code path that accepts a host or an asset as a target.
- The console shows a persistent "authorised testing only" notice.

These are technical guardrails, not a substitute for authorisation. Responsibility for having it stays with the operator.
