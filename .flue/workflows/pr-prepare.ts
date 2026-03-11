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
