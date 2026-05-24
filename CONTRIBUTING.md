# Contributing

We welcome contributions to the Agent Coordination Substrate specification.

## Types of Contributions

- **Spec clarifications** — Improved normative language, edge case documentation
- **Schema updates** — New fields, validation improvements, version bumps
- **Conformance scenarios** — Additional test expectations
- **Reference implementation** — Bug fixes, new store implementations
- **Comparison updates** — New alternatives, capability changes in existing alternatives

## Process

1. Open an issue describing the change
2. Fork and create a branch
3. Make changes with clear commit messages
4. Submit a pull request referencing the issue

## Naming Conventions

This spec uses engineering-native vocabulary. See the naming table in `advisory/SPEC.md` and `enforcement/SPEC.md`. Do not introduce biology-metaphor terminology (pheromone, scent, trail, emit, sniff) into normative text. The theoretical lineage is acknowledged in `docs/theory.md`.

## Normative Language

Spec documents use RFC 2119 keywords (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY). Use these precisely and only when specifying behavioral requirements.

## Style

- Keep prose concise
- One concept per section
- Examples accompany every normative requirement
- Schemas validate against JSON Schema Draft 2020-12
