import { github, githubBody, type ProxyFactory, type ProxyPolicy } from "@flue/client/proxies";
import * as v from "valibot";

const githubProxy = github({
  policy: {
    base: "allow-read",
    allow: [
      { method: "POST", path: "/graphql", body: githubBody.graphql() },
      { method: "POST", path: "/*/git-upload-pack" },
      { method: "POST", path: "/*/git-receive-pack" }
    ]
  }
});

const useSandbox =
  (globalThis as typeof globalThis & { process?: { env?: Record<string, string | undefined> } }).process?.env
    ?.FLUE_USE_SANDBOX === "true";

export const proxies =
  useSandbox
    ? {
        github: githubProxy,
        // Sandbox mode cannot reuse host-side auth/config directly.
        opencode: opencodeZen()
      }
    : {
        // Host mode behaves like local opencode: the runner's own auth/config selects the provider.
        github: githubProxy
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

function opencodeZen(opts?: { policy?: string | ProxyPolicy }): ProxyFactory<{ apiKey: string }> {
  const factory = ({ apiKey }: { apiKey: string }) => ({
    name: "opencode-zen",
    target: "https://api.opencode.ai",
    headers: {
      authorization: `Bearer ${apiKey}`,
      host: "api.opencode.ai"
    },
    policy: opts?.policy ?? "allow-all",
    isModelProvider: true,
    providerConfig: {
      providerKey: "opencode",
      options: {
        apiKey: "sk-dummy-value-real-key-injected-by-proxy"
      }
    }
  });

  factory.secretsMap = { apiKey: "OPENCODE_API_KEY" };
  factory.proxyName = "opencode";
  return factory;
}
