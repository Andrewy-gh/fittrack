# Hypertrophy assessment investigation

2026-09-22. Research follow-up, not an implemented or finally approved numeric policy. The user endorsed extending the strength-policy direction and authorized investigation. Existing strength behavior remains unchanged.

## Conclusion

Proceed with an assessment-specific muscle mapping and transparent weekly counts. A fixed pass/fail threshold for fractional counts is not established by this investigation. Do not silently combine a threshold from one counting convention with totals from another.

Carry forward Analytics placement, deterministic calculations, general-reference framing, completed local weeks, counts-only current weeks, visible uncertainty, and separation from individualized prescriptions. These decisions do not need reopening.

## Research findings and access

1. [Pelland et al., published online 2025, Sports Medicine 2026](https://doi.org/10.1007/s40279-025-02344-w), sections 2.5, 4.6 and Table 1: direct plus half of indirect sets had strongest relative model support. Classification depends on the measured muscle, not a generic secondary label. The authors acknowledge judgment in mapping. No clear hypertrophy plateau was identified. [Published-paper PDF inspected](https://www.fisiologiadelejercicio.com/wp-content/uploads/2025/12/The-Resistance-Training-Dose-Response.pdf); OSF supplement access failed, and extracted data/model code were not independently reproduced.
2. [ACSM 2026 position stand](https://pmc.ncbi.nlm.nih.gov/articles/PMC12965823/), hypertrophy results and Table 6: supports higher weekly volume, including a reference of at least ten sets per muscle group. This is general healthy-adult guidance, not an individual growth threshold. The paper does not specify that this reference means ten fractional sets. Its discussion describes diminishing returns around 18–20 sets; that should not become an app ceiling, particularly given the more qualified conclusion of the cited original analysis. Main full text inspected; supplementary review tables not independently audited.
3. [Schoenfeld, Ogborn and Krieger, 2017](https://pubmed.ncbi.nlm.nih.gov/27433992/): the three volume categories were fewer than five, five through nine, and ten or more weekly sets. Their categorical comparison was a trend (P=.074), while continuous volume and within-study higher/lower comparisons supported a dose relationship. This does not identify a biological cutoff at ten. [Author-uploaded full text](https://www.researchgate.net/publication/305455324_Dose-response_relationship_between_weekly_resistance_training_volume_and_increases_in_muscle_mass_A_systematic_review_and_meta-analysis) inspected for coding/statistical methods; no fractional conversion was established there.
4. [Kubo et al., 2019](https://doi.org/10.1007/s00421-019-04181-y): a small ten-week squat study in 17 men found no significant hamstring-volume change, although other measured muscles grew. This challenges automatic hamstring credit from a squat's catalog label; it does not prove every squat has zero hamstring stimulus. [Published-paper PDF inspected](https://highfit.com.br/wp-content/uploads/2019/06/KEITARO-KUBO-2019.pdf).

## Candidate mappings against the actual catalog

These are scoped candidates derived from Pelland Table 1, not a complete muscle map or proof that variants are identical:

| FitTrack catalog movement | Candidate outcome and role |
| --- | --- |
| Medium-grip barbell bench press | Chest direct; triceps indirect |
| Triceps pushdown | Triceps direct |
| Barbell/hammer curl | Elbow flexors direct |
| Bent-over barbell row, seated cable row, wide-grip pulldown | Elbow flexors indirect |
| Full squat, leg press, dumbbell lunge, leg extension | Quadriceps group direct |
| Lying leg curl | Hamstrings direct |
| Incline dumbbell press, dumbbell shoulder press | Triceps indirect |

Each catalog-to-paper variant match still needs an explicit rationale in the implementation mapping. Group results must not imply uniform stimulus across every constituent muscle.

Remaining catalog movements/roles need separate mapping review: dumbbell bench press, pushups, one-arm dumbbell row, pullups, Romanian deadlift, conventional deadlift, crunches, plank, lateral raise, and both calf raises. Unlisted roles on the candidate movements also remain unresolved. In particular, do not derive a deadlift's muscle credits by applying 0.5 to its eight catalog secondary labels. A missing research-table entry means unreviewed here, not no training effect.

## Proposed product rules

These are engineering/product recommendations, not additional findings from the papers:

- Store a versioned assessment mapping separate from the source catalog. Preserve user exercise identities, optional classification, and existing raw evidence semantics.
- Retain direct and indirect counts separately. Any weighted display must identify itself as an estimate and expose its components. Do not call it effective sets or measured stimulus.
- Track each muscle's mapped, unresolved, excluded, and invalid contributions. A catalog link alone does not establish assessment coverage. Never add muscle totals together into a whole-body set score.
- Show known contributions when coverage is partial. An unclassified exercise can potentially affect any muscle; a reviewed exclusion can narrow that uncertainty. Missing evidence must not generate a below-reference conclusion. Numerical comparison, if later enabled, must explicitly concern mapped logged data.
- Preserve existing assessment row validation and week boundaries. Current periods remain counts-only. No logs remain unknown. Invalid attributable rows prevent a confident comparison for the affected muscle; unassignable invalid rows require broader uncertainty.
- Do not infer effort, range of motion, deload intent, or real-world logging completeness. The current set request records reps, optional weight, and working/warmup type; the profile records goals and constraints but no structured RIR/RPE. Session Metrics do not fill that gap.
- No additional profile fields are necessary for an explicitly limited count summary. Isometric and unilateral records need explicit handling: a plank must not be treated as an ordinary dynamic repetition set, and no side-doubling should be guessed from names or repetitions.

## Worked acceptance examples

| Example | Proposed behavior |
| --- | --- |
| Six reviewed bench sets plus four pushdown sets | Triceps: four direct and six indirect contributions; weighted arithmetic is seven. Keep the ten unweighted contributions distinguishable; do not compare seven with a threshold defined in unweighted units. |
| Above, plus an unclassified custom exercise | Retain known contributions, show incomplete attribution, and suppress a whole-muscle below-reference verdict. |
| Six squat sets with catalog hamstring secondary label | Do not automatically report three weighted hamstring sets. Mapping requires outcome-specific justification. |
| Ten mapped contributions in an unfinished week | Facts so far; no comparison badge. |
| A muscle absent from all known mappings | No supported mapped evidence; do not assert the user did no training for it. |
| Classification changed after logging | Recalculate using current links, consistent with the evidence service; include mapping/catalog versions. Do not promise immutable historical assessments. |

## Recommended next implementation boundary

Build the reviewed mapping and contribution explanation first, with the existing Analytics period and uncertainty behavior. A neutral research-reference explanation can accompany the counts. Keep a numeric hypertrophy reference badge gated until the chosen metric and benchmark are justified together.

The remaining policy decision is precise: either ship this informative count view first, or explicitly adopt a numeric product convention with its limitations. This investigation does not establish a scientifically validated ten-fractional-set cutoff. It also does not establish a ten-to-twenty optimal range, growth prediction, automatic workload increase, or individualized target.

Validation for this investigation: inspected the current catalog, profile and workout request models, evidence/assessment contracts, source methods and limitations; checked worked arithmetic and document whitespace. No application code changed, no tests were needed, and no commit, PR, or deployment was performed.
