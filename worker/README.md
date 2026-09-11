# KerenScope hosted demo LLM endpoint

A tiny Cloudflare Worker that lets anyone run the KerenScope CLI without
configuring an LLM key: the CLI falls back to this endpoint (rate-limited)
when no `KERENSCOPE_LLM_*` configuration exists.

- **Model whitelist**: `@cf/zai-org/glm-5.3-flash` (and `glm-5.3` fallback) only
- **Rate limit**: 10 requests/minute per IP (Cloudflare rate limiting binding)
- **max_tokens capped** at 16,384 per request; payloads capped at 1 MB
- **No credentials in the repo**: the upstream Workers AI token is a Worker
  Secret (`CF_API_KEY`), set at deploy time

The endpoint only serves requests carrying the `X-KerenScope-Demo: 1` header
(the CLI always sends it), which keeps casual scanners out.

## Deploy

```bash
cd worker
npx wrangler deploy
echo "$CLOUDFLARE_API_KEY" | npx wrangler secret put CF_API_KEY
```

## Disable

```bash
npx wrangler delete   # or pause it from the Cloudflare dashboard
```

Once the endpoint is gone, the CLI automatically asks for the user's own
OpenAI-compatible key instead — the product keeps working.
