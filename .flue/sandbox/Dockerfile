FROM node:22-bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       ca-certificates curl wget git jq \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g pnpm

RUN curl -fsSL https://opencode.ai/install | bash \
    && cp /root/.opencode/bin/opencode /usr/local/bin/opencode

RUN (type -p wget >/dev/null || (apt-get update && apt-get install wget -y)) \
    && mkdir -p -m 755 /etc/apt/keyrings \
    && wget -nv -O /tmp/gh.gpg https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    && cat /tmp/gh.gpg | tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
    && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
       | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update && apt-get install gh -y \
    && rm -rf /var/lib/apt/lists/*

RUN git config --system --add safe.directory '*'

EXPOSE 48765
CMD ["opencode", "serve", "--port", "48765", "--hostname", "0.0.0.0"]
