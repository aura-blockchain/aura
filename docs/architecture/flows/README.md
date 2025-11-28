# Flow Diagrams

Use these diagrams to visualize core protocol interactions. Source files (PlantUML, Mermaid, Draw.io) should be committed alongside exported PNG/SVG.

## Available Diagrams

1. **IR Completion Flow** – `ir-completion.puml` + `ir-completion.svg`
2. **VC Mint & Revocation** – `vc-mint-revoke.puml` + `vc-mint-revoke.svg`
3. **Verifier Proof Flow** – `verifier-proof-flow.puml` + `verifier-proof-flow.svg`
4. **ZK Governance Vote** – `zk-governance-vote.puml` + `zk-governance-vote.svg`
5. **Assistant Slashing & Freeze** – `assistant-slashing.puml` + `assistant-slashing.svg`

## Exporting

Use PlantUML to regenerate assets any time a `.puml` changes:

```
java -jar plantuml.jar -tsvg docs/architecture/flows/*.puml
```

Commit both the source and exported SVG so docs consumers can embed diagrams without rebuilding.
