export type PublishCommand = {
  branch: string;
  sha: string;
};

export function preparedBranchName(pullNumber: number, headSHA: string, prefix = "agent/prepared"): string {
  return `${prefix.replace(/\/$/, "")}/pr-${pullNumber}-${headSHA.slice(0, 7)}`;
}

export function buildPublishCommand(branch: string, sha: string): string {
  return `/pr-agent publish branch=${branch} sha=${sha}`;
}

export function parsePublishCommand(commentBody: string): PublishCommand | null {
  const match = commentBody.match(/\/pr-agent\s+publish\s+branch=([^\s]+)\s+sha=([0-9a-fA-F]{7,64})/);
  if (!match) {
    return null;
  }

  return {
    branch: match[1],
    sha: match[2]
  };
}
