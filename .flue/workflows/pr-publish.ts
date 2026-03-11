import type { FlueClient } from "@flue/client";
import { parsePublishCommand } from "./commands.ts";
import { mergeResolutionSchema, proxies, repoSlug, shellQuote } from "./shared.ts";

export { proxies };

type Args = {
  owner: string;
  repo: string;
  pullNumber: number;
  actor: string;
  commentBody: string;
};

export default async function publish(flue: FlueClient, args: Args) {
  const repository = repoSlug(args.owner, args.repo);
  const command = parsePublishCommand(args.commentBody);
  if (!command) {
    return;
  }

  await configureGit(flue);

  const prView = await flue.shell(
    `gh pr view ${args.pullNumber} --repo ${shellQuote(repository)} --json number,title,url,headRefName,headRefOid,baseRefName,isCrossRepository`
  );
  const pr = JSON.parse(prView.stdout);
  const headRefName = String(pr.headRefName);

  if (pr.headRefOid !== command.sha) {
    await postComment(
      flue,
      repository,
      args.pullNumber,
      [
        "## PR agent publish",
        "",
        `Refused to publish because the PR head changed from \`${command.sha}\` to \`${pr.headRefOid}\`.`,
        "Run `/pr-agent prepare` again to generate a fresh prepared branch."
      ].join("\n")
    );
    throw new Error(`stale prepared branch: expected ${command.sha}, got ${pr.headRefOid}`);
  }

  if (pr.isCrossRepository) {
    await postComment(
      flue,
      repository,
      args.pullNumber,
      [
        "## PR agent publish",
        "",
        "Fork-based pull requests are not merged automatically by this GitHub Actions setup."
      ].join("\n")
    );
    return;
  }

  await flue.shell(`gh pr checkout ${args.pullNumber} --repo ${shellQuote(repository)}`);
  await flue.shell(`git checkout -B ${shellQuote(headRefName)} ${shellQuote(`origin/${headRefName}`)}`);
  await flue.shell(`git fetch origin ${shellQuote(command.branch)}`);

  const merge = await flue.shell(`git merge --no-ff --no-edit ${shellQuote(`origin/${command.branch}`)}`);
  if (merge.exitCode !== 0) {
    const resolution = await flue.skill(".flue/skills/resolve-merge-conflicts.md", {
      args: {
        pullRequest: pr,
        preparedBranch: command.branch,
        mergeStdout: merge.stdout,
        mergeStderr: merge.stderr
      },
      result: mergeResolutionSchema
    });

    if (!resolution.resolved) {
      await postComment(
        flue,
        repository,
        args.pullNumber,
        [
          "## PR agent publish",
          "",
          "Merge conflicts were detected and the conflict-resolution skill did not finish safely.",
          "Please inspect the prepared branch manually.",
          "",
          `Prepared branch: \`${command.branch}\``
        ].join("\n")
      );
      throw new Error("merge conflict resolution failed");
    }

    await flue.shell("git add -A");
    await flue.shell(`git commit -m ${shellQuote(`chore(pr-agent): resolve merge conflicts for PR #${args.pullNumber}`)}`);
  }

  await flue.shell(`git push origin HEAD:${shellQuote(headRefName)} --force-with-lease`);

  await postComment(
    flue,
    repository,
    args.pullNumber,
    [
      "## PR agent publish",
      "",
      `Prepared branch \`${command.branch}\` was merged into \`${headRefName}\`.`
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
