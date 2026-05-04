---
name: git-open-pull
description: "Create a GitHub Pull Request using git-open-pull. Use when: opening a PR, submitting code for review, creating a draft PR, publishing a branch as a pull request. git-open-pull creates an issue, renames the branch to include the issue number, pushes the branch, and converts the issue to a PR."
argument-hint: "Optionally specify a title, labels, description file, or whether to create as a draft"
---

# git-open-pull Skill

`git-open-pull` converts the current git branch into a GitHub pull request by:
1. Opening or creating a GitHub issue (with title, description, and labels)
2. Renaming the local branch to embed the issue number (e.g. `my-feature` → `my-feature_42`)
3. Pushing the renamed branch and converting the issue into a pull request

Use `git-open-pull` instead of `gh pr create` when working in a repository that tracks work via GitHub Issues and encodes the issue number in the branch name.

## When to Use

- The user wants to open a PR for their current branch
- The user has finished a feature or fix and wants to submit it for review
- The user wants to create a draft PR to share work in progress
- The repository uses the convention of embedding issue numbers in branch names

## Procedure

### 1. Check Configuration

Run `git-open-pull --help` first. If required configuration is missing, the help output will list any pre-req information needed (e.g. GitHub username, token, destination account/repo). Set these values via `git config` or environment variables as described in the help output before proceeding.

### 2. Check for Uncommitted Changes

Before running, ensure all changes are committed. `git-open-pull` does not commit changes — it only renames the branch, pushes it, and opens the PR.

### 3. Do Not Push the Branch First

**Do not run `git push` before `git-open-pull`.** The tool renames the local branch to include the issue number before pushing. If you push manually beforehand the remote branch name will not match and the tool will fail.

### 4. Inspect Labels (if using --labels)

Run `git-open-pull --list-labels` to see the exact label names valid for this repository before passing `--labels`.

### 5. Prepare PR Details

Write a good title and description:

**Title**: Use imperative mood, keep it under 72 characters, and describe *what* the PR does (e.g. `Add retry logic for failed API requests`).

**Description file**: Write to a temp file and pass via `--description-file`. Include:
- A short summary of what changed and why
- Any relevant issue references (e.g. `Closes #123` — note: `git-open-pull` converts the issue to a PR, so the issue number is already linked)
- Notable implementation decisions useful for the reviewer

### 6. Run git-open-pull

```sh
git-open-pull \
  --interactive=false \
  --title "Fix null pointer in login flow" \
  --description-file /tmp/pr-description.txt \
  --labels bug \
  --draft
```

### 7. Confirm Result

On success the tool prints the GitHub issue/PR URL:
```
https://github.com/owner/repo/issues/42
```

Report this URL to the user.

## Flags

| Flag | Description |
|------|-------------|
| `--title` | PR / issue title (required with `--interactive=false`) |
| `--description-file` | Path to a file whose contents become the PR description |
| `--labels` | Comma-separated label names (use `--list-labels` to enumerate valid values) |
| `--draft` | Open the PR in draft mode (default: true) |
| `--interactive` | Set to `false` for non-interactive/agent use |
| `--list-labels` | Print all repository labels and exit |
| `--skill` | Print this skill document and exit |
| `--version` | Print the version and exit |

## Best Practices

### Titles
- Use the imperative mood: `Fix`, `Add`, `Update`, `Remove`, `Refactor` — not `Fixed`, `Adding`, etc.
- Be specific: `Fix null pointer in user login flow` beats `Fix bug`.
- Keep it under 72 characters so it displays cleanly in GitHub and email notifications.

### Descriptions
- Start with a one-sentence summary.
- Explain *why* the change is needed, not just *what* it does — reviewers benefit from context.
- If the change is large, add a brief list of the main files or components touched.

### Draft PRs
- Use `--draft` when the code is not yet ready for formal review (e.g. work in progress, awaiting feedback on approach, CI not yet passing).
- Draft PRs are visible to collaborators but will not show as review-requested until marked ready.
