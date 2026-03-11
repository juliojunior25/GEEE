# GEEE Merge Conflict Resolution Skill

You are resolving merge conflicts between:

- the current PR head branch
- a prepared agent branch created by the `prepare` workflow

## Goal

Resolve conflicts conservatively and preserve the PR author's intent first.

## Rules

1. Keep the PR author's changes as the source of truth.
2. Re-apply only the prepared changes that are still clearly valid.
3. If a prepared change is no longer appropriate, drop it.
4. Prefer the smallest safe resolution.
5. Run the lightest relevant validation after resolving conflicts.
6. If you cannot resolve safely, stop and return failure.

## Output

Return JSON with this exact shape:

```json
{
  "resolved": true,
  "summary": "Short description of how the conflicts were resolved."
}
```

If safe resolution is not possible:

```json
{
  "resolved": false,
  "summary": "Short explanation of why manual intervention is required."
}
```
