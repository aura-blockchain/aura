# Validator APR Scenarios

Companion to `validator-apr-scenarios.csv`, which models validator/ delegator returns under different emission eras. Inputs derive from `emissions-schedule.*` (minted AEQ per year) and assume validator rewards receive 60% of each year’s pool.

## Key Inputs
- **AEQ price reference:** $0.50 (adjust per market).
- **Total bonded stake:** varies per scenario (400M–550M AEQ) to reflect adoption.
- **Uptime factor:** linear multiplier on rewards; validators missing votes earn proportionally less.
- **Self-bond %:** share of validator stake owned by the operator (default 7–12%). Impacts APR via self rewards + commission take.
- **Commission:** portion of delegator rewards paid to the validator operator.

## Scenario Takeaways
- **Year 1 Top Validator:** With 5M AEQ stake, 99% uptime, 5% commission, operator earns ~25.8% APR on self-bond and delegators see ~16.9% APR thanks to front-loaded emissions.
- **Year 1 Mid Validator:** Smaller 2M stake and 3% commission still yields similar APRs provided uptime remains high; demonstrates accessibility for smaller operators.
- **Year 5 Tail:** Emissions drop to 40M AEQ/year, cutting validator APR to ~6.8% (self-bond) and delegator APR to ~4.5% absent fee revenue, reinforcing the need for transaction fee markets.
- **Year 8 Tail:** Long-term emissions (16M AEQ) plus higher commission (7%) shrink delegator APR to ~1.5% unless fee or MEV income supplements rewards.

## Next Steps
1. Integrate actual fee revenue assumptions (per-block gas, PoI verifier fees) to show holistic APRs.
2. Build a notebook that sweeps commission, uptime, and price inputs to create sensitivity charts for governance.
3. Tie validator APR outputs into the broader emissions simulations to validate sustainability targets.
