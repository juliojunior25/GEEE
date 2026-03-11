# GEEE PR Prepare Skill

You are preparing a small, safe follow-up change for the current pull request in the GEEE repository.

GEEE is a Go project focused on extraction and enrichment pipelines, plugins, low resource usage, and predictable behavior.

## Inputs

- `pullRequest`: metadata from `gh pr view`
- `diffStat`: diff summary between the PR head and base branch
- `trigger`: what started this run
- `actor`: who started it
- `commentBody`: original slash command, if any

## Preferred kinds of improvements

Choose only one focused improvement area:

1. add or tighten small tests around the new behavior
2. fix obvious documentation drift in README, examples, or config comments
3. add missing validation or error handling adjacent to the new code
4. small Go cleanup that makes the new change safer or clearer

## Constraints

- Stay close to the pull request scope.
- Do not do broad refactors.
- Do not introduce new product behavior unless clearly implied by the PR.
- Prefer targeted `go test ./...` or the smallest relevant test command after changes.
- If there is no clearly safe improvement, return no changes.

## Output

Return JSON with this exact shape:

```json
{
  "changesRequired": true,
  "summary": "One short sentence about what was prepared.",
  "reasoning": "Why this follow-up change is justified for this pull request.",
  "filesTouched": ["path/one", "path/two"]
}
```

If no safe change is justified:

```json
{
  "changesRequired": false,
  "summary": "No safe follow-up change was identified.",
  "reasoning": "Brief explanation.",
  "filesTouched": []
}
```
