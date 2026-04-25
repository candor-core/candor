# Candor Founding Contributor Program

> This is not a commercial tier. There is no pricing, no discount, no access gate.
> Founding contributors earn governance standing. That is the only thing on offer —
> and at this stage of the language, it is the most valuable thing available.

---

## What "founding contributor" means

A founding contributor is an organization or individual who puts Candor into real use
during the formative period — before the surface hardens, before the idioms calcify —
and feeds that experience back into Core.

**Contribution is not code. It is signal.**

Signal means:
- Use cases that expose gaps in the language before decisions harden around them
- Idioms that become canonical because a real codebase validated them
- Problems that Core can still fix — rather than document as limitations

An organization that runs a Candor pipeline in production and files an RFC about what
`?` needs to do differently at a module boundary is doing more for the language than
any pull request.

---

## What founding contributors receive

### RFC standing

Founding contributors have standing to propose, comment on, and ratify RFCs. This is
the formal mechanism by which the language changes. Core does not accept amendments
without founding contributor ratification. The council is small by design — each voice
carries real weight.

### Amendment ratification

The Candor axioms are frozen (see `docs/spec/L0-AXIOMS.md`). The surface layer is not.
When a decision about syntax, operator semantics, or standard library shape comes to
a vote, founding contributors are the voting body. Their use cases inform the options;
their ratification makes decisions durable.

### Use cases cited in the specification

The problems a founding contributor brought to Core — the actual production requirements,
the edge cases, the things the language got wrong — are cited in the relevant spec sections.
This is not acknowledgment for its own sake. It is the record of why a decision was made,
which is the only thing that protects a decision from being undone later by someone who
wasn't there.

### Permanent attribution in Core

"Candor Core by Scott W. Corley, with founding contributors [names]" is in the history.
That attribution is non-transferable. An organization that shapes the `?` operator at its
inflection point has a codebase that reads like the language was designed for them —
because it partly was.

---

## Why contribution beats payment at this stage

Architectural influence over a language at its inflection point is asymmetric. A pricing
discount is symmetric — anyone can buy it later. The organizations that shape `?` and
pure arena semantics now will have idioms in production that the language then validates
and extends. That is a compounding position.

Paying for a license after the language ships buys access. Contributing before it ships
buys authorship. These are not the same thing.

---

## What is not on offer

- No commercial licensing terms, no SLA, no support contract
- No exclusive features, no private branch
- No equity in any entity (there is none)
- No guarantee of roadmap direction — the council is the mechanism for influence,
  not a promise that any specific feature will land

If your organization needs vendor guarantees, Candor is not the right fit yet. If your
organization wants to be the reason a language decision went one way rather than another,
read on.

---

## The governance model

```
Layer 1 — Founding contributors
  RFC standing, amendment ratification, use cases in spec
  Earned by: production use + signal back to Core during the formative period

Layer 2 — Ecosystem builders
  Packages, integrations, tooling built on the stable surface
  Earned by: shipping something real that other Candor users depend on

Layer 3 — Self-improvement
  Community + AI loop closes gaps; compiler improves itself
  Earned by: participation in the corpus, benchmarks, and open issues
```

The three layers are not a hierarchy of importance. They are a division of labor.
Layer 1 shapes what the language is. Layer 2 shapes what the ecosystem becomes.
Layer 3 shapes how quickly gaps close.

---

## Design constraint: low author bandwidth

Scott's primary work is elsewhere. The governance structure is designed to carry itself:

- Founding contributors hold amendment ratification — Core does not change over their
  objection
- The council is small (target: 5–12 founding contributors) so consensus is achievable
- The methodology self-propagates through the people who take it seriously — the same
  way Haskell's purity culture spread through developers who understood what it was for

This is not a governance structure that requires constant attention from a single person.
It is a structure that stays durable at low bandwidth because each participant has a
real stake in the decisions they are ratifying.

---

## How to become a founding contributor

Founding contributor standing is not applied for. It is established through demonstrated
use and engagement during the formative period.

If your organization is using Candor in production, or is evaluating it seriously enough
to have opinions about what it gets wrong, reach out directly:

**Scott W. Corley** — scott.corley.g1@gmail.com

The conversation worth having: what is your use case, what did the language do wrong for
it, and what would you need it to do differently. That conversation is the contribution.

---

*Candor Core — Apache-2.0. Developed by Scott W. Corley.*
