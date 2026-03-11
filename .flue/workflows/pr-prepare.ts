import type { FlueClient } from "@flue/client";
import { buildPublishCommand, preparedBranchName } from "./commands.ts";
import { preparedResultSchema, proxies, repoSlug, shellQuote } from "./shared.ts";

export { proxies };

type Args = {
  owner: string;
  repo: string;
  pullNumber: number;
  trigger: string;
  actor: string;
  commentBody?: string;
  prepareBranchPrefix?: string;
};

export default async function prepare(flue: FlueClient, args: Args) {
  const repository = repoSlug(args.owner, args.repo);

  await configureGit(flue);

  const prView = await flue.shell(
    `gh pr view ${args.pullNumber} --repo ${shellQuote(repository)} --json number,title,body,url,headRefName,headRefOid,baseRefName,author,isCrossRepository,files`
  );
  const pr = JSON.parse(prView.stdout);

  if (pr.isCrossRepository) {
    await postComment(
      flue,
      repository,
      args.pullNumber,
      [
        "## PR agent prepare",
        "",
        "This setup currently supports pull requests whose head branch lives in the same repository.",
        "Fork-based pull requests are intentionally not pushed automatically."
      ].join("\n")
    );
    return;
  }

  await flue.shell(`gh pr checkout ${args.pullNumber} --repo ${shellQuote(repository)}`);
  await flue.shell(`git fetch origin ${shellQuote(pr.baseRefName)}`);
  const diff = await flue.shell(`git diff --stat origin/${pr.baseRefName}...HEAD`);

  const validation = await flue.shell("go test ./...", { timeout: 30 * 60 * 1000 });
  if (validation.exitCode !== 0) {
    const suggestion = deriveValidationSuggestion(validation.stdout, validation.stderr);
    await postComment(
      flue,
      repository,
      args.pullNumber,
      [
        "## PR agent prepare",
        "",
        "Blocking issue detected before any follow-up change was prepared.",
        "",
        "Validation command:",
        "",
        "`go test ./...`",
        "",
        "Output excerpt:",
        "",
        "```text",
        summarizeValidationOutput(validation.stdout, validation.stderr),
        "```",
        "",
        "Possible correction:",
        "",
        suggestion
      ].join("\n")
    );
    throw new Error("blocking validation failed");
  }

  const result = await flue.skill(".flue/skills/pr-prepare.md", {
    args: {
      trigger: args.trigger,
      actor: args.actor,
      commentBody: args.commentBody ?? "",
      pullRequest: pr,
      diffStat: diff.stdout
    },
    result: preparedResultSchema
  });

  const status = await flue.shell("git status --porcelain");
  if (!result.changesRequired || status.stdout.trim().length === 0) {
    await postComment(
      flue,
      repository,
      args.pullNumber,
      [
        "## PR agent prepare",
        "",
        `Summary: ${result.summary}`,
        "",
        `Reasoning: ${result.reasoning}`,
        "",
        "No code changes were prepared.",
        "To run the agent again, comment `/pr-agent prepare`."
      ].join("\n")
    );
    return;
  }

  const preparedBranch = preparedBranchName(args.pullNumber, pr.headRefOid, args.prepareBranchPrefix);
  await flue.shell(`git checkout -B ${shellQuote(preparedBranch)}`);
  await flue.shell("git add -A");
  await flue.shell(`git commit -m ${shellQuote(`chore(pr-agent): prepare changes for PR #${args.pullNumber}`)}`);
  await flue.shell(`git push origin HEAD:${shellQuote(preparedBranch)} --force-with-lease`);

  const compareURL = `https://github.com/${repository}/compare/${pr.headRefName}...${preparedBranch}`;
  const publishCommand = buildPublishCommand(preparedBranch, pr.headRefOid);

  await postComment(
    flue,
    repository,
    args.pullNumber,
    [
      "## PR agent prepare",
      "",
      `Summary: ${result.summary}`,
      "",
      `Reasoning: ${result.reasoning}`,
      "",
      `Prepared branch: \`${preparedBranch}\``,
      `Compare diff: ${compareURL}`,
      "",
      "To merge the prepared changes back into the PR branch, comment:",
      "",
      `\`${publishCommand}\``
    ].join("\n")
  );
}

async function configureGit(flue: FlueClient) {
  await flue.shell(`git config user.name "github-actions[bot]"`);
  await flue.shell(`git config user.email "41898282+github-actions[bot]@users.noreply.github.com"`);
}

async function postComment(flue: FlueClient, repository: string, pullNumber: number, body: string) {
  await flue.shell(`gh pr comment ${pullNumber} --repo ${shellQuote(repository)} --body-file -`, { stdin: body });
}

function summarizeValidationOutput(stdout: string, stderr: string): string {
  const merged = [stdout.trim(), stderr.trim()].filter(Boolean).join("\n");
  const lines = merged.split("\n").filter(Boolean);
  const excerpt = lines.slice(-30).join("\n").trim();
  return excerpt.length > 0 ? excerpt : "Validation failed without stdout/stderr output.";
}

function deriveValidationSuggestion(stdout: string, stderr: string): string {
  const merged = [stdout.trim(), stderr.trim()].filter(Boolean).join("\n");

  const undefinedSymbolMatch = merged.match(/([^\s:]+\.go):\d+:\d+:\s+undefined:\s+([^\s]+)/);
  if (undefinedSymbolMatch) {
    const [, file, symbol] = undefinedSymbolMatch;
    return [
      `- Open \`${file}\` and either define \`${symbol}\` in the same package or remove the reference if it was accidental.`,
      "- Re-run `go test ./...` locally after the edit to confirm the package compiles again."
    ].join("\n");
  }

  const unusedImportMatch = merged.match(/([^\s:]+\.go):\d+:\d+:\s+\"([^\"]+)\"\s+imported and not used/);
  if (unusedImportMatch) {
    const [, file, importPath] = unusedImportMatch;
    return [
      `- Remove the unused import \`${importPath}\` from \`${file}\`, or start using it in code if it is required.`,
      "- Re-run `go test ./...` to verify the compile error is gone."
    ].join("\n");
  }

  const unusedVarMatch = merged.match(/([^\s:]+\.go):\d+:\d+:\s+declared and not used:\s+([^\s]+)/);
  if (unusedVarMatch) {
    const [, file, variable] = unusedVarMatch;
    return [
      `- In \`${file}\`, remove the unused variable \`${variable}\` or make the code use it before returning.`,
      "- Re-run `go test ./...` after the cleanup."
    ].join("\n");
  }

  const missingModuleMatch = merged.match(/no required module provides package\s+([^\s;]+);/);
  if (missingModuleMatch) {
    const [, pkg] = missingModuleMatch;
    return [
      `- Add the missing module for \`${pkg}\` with \`go get ${pkg}\`, or remove the import if it should not be there.`,
      "- Commit any resulting `go.mod` and `go.sum` updates, then re-run `go test ./...`."
    ].join("\n");
  }

  const failedTestMatch = merged.match(/--- FAIL: ([^(\\s]+)/);
  if (failedTestMatch) {
    const [, testName] = failedTestMatch;
    return [
      `- Start by reproducing the failing test with \`go test ./... -run ${testName}\` to narrow the problem.`,
      "- Fix the assertion or production code causing the failure, then re-run the full test suite."
    ].join("\n");
  }

  return [
    "- Run `go test ./...` locally and fix the first compile or test error shown in the output excerpt above.",
    "- After the fix, push the branch or comment `/pr-agent prepare` again to re-run validation."
  ].join("\n");
}
