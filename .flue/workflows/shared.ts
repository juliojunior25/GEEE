import { github, githubBody } from "@flue/client/proxies";
import * as v from "valibot";

export const proxies = {
  github: github({
  policy: {
    base: "allow-read",
    allow: [
      { method: "POST", path: "/graphql", body: githubBody.graphql() },
      { method: "POST", path: "/*/git-upload-pack" },
      { method: "POST", path: "/*/git-receive-pack" }
    ]
  }
  })
};

export const preparedResultSchema = v.object({
  changesRequired: v.boolean(),
  summary: v.string(),
  reasoning: v.string(),
  filesTouched: v.optional(v.array(v.string()))
});

export const mergeResolutionSchema = v.object({
  resolved: v.boolean(),
  summary: v.string()
});

export function shellQuote(value: string): string {
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

export function repoSlug(owner: string, repo: string): string {
  return `${owner}/${repo}`;
}
