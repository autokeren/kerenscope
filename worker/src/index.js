const ALLOWED_MODELS = new Set([
  "@cf/zai-org/glm-5.3-flash",
  "@cf/zai-org/glm-5.3",
]);
const MAX_TOKENS_CAP = 16384;
const MAX_BODY_BYTES = 1024 * 1024;

const buckets = new Map();
const WINDOW_MS = 60_000;
const PER_IP_LIMIT = 30;

function allowRequest(ip) {
  const now = Date.now();
  if (buckets.size > 10000) buckets.clear();
  let bucket = buckets.get(ip);
  if (!bucket || now - bucket.start > WINDOW_MS) {
    bucket = { start: now, count: 0 };
    buckets.set(ip, bucket);
  }
  bucket.count++;
  return bucket.count <= PER_IP_LIMIT;
}

function json(status, obj) {
  return new Response(JSON.stringify(obj), {
    status,
    headers: { "content-type": "application/json" },
  });
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (request.method === "GET" && (url.pathname === "/" || url.pathname === "/v1")) {
      return json(200, {
        service: "kerenscope-llm-demo",
        models: [...ALLOWED_MODELS],
        note: "free hosted demo for the KerenScope CLI — bring your own key any time",
      });
    }
    if (request.method !== "POST" || url.pathname !== "/v1/chat/completions") {
      return json(404, { error: "not found" });
    }
    if (request.headers.get("x-kerenscope-demo") !== "1") {
      return json(401, { error: "this demo endpoint only serves the KerenScope CLI" });
    }
    const ip = request.headers.get("cf-connecting-ip") || "anon";
    if (!allowRequest(ip)) {
      return json(429, {
        error: "rate limit reached — wait a moment or configure your own LLM key (KERENSCOPE_LLM_API_KEY)",
      });
    }
    let body;
    try {
      const raw = await request.text();
      if (raw.length > MAX_BODY_BYTES) return json(413, { error: "payload too large" });
      body = JSON.parse(raw);
    } catch {
      return json(400, { error: "invalid JSON body" });
    }
    const model = body.model || "@cf/zai-org/glm-5.3-flash";
    if (!ALLOWED_MODELS.has(model)) {
      return json(400, { error: `model not available on this demo endpoint: ${model}` });
    }
    body.model = model;
    body.stream = false;
    if (!body.max_tokens || body.max_tokens > MAX_TOKENS_CAP) {
      body.max_tokens = MAX_TOKENS_CAP;
    }
    const upstream = await fetch(
      `https://api.cloudflare.com/client/v4/accounts/${env.CF_ACCOUNT_ID}/ai/v1/chat/completions`,
      {
        method: "POST",
        headers: {
          "content-type": "application/json",
          authorization: `Bearer ${env.CF_API_KEY}`,
        },
        body: JSON.stringify(body),
      }
    );
    const text = await upstream.text();
    return new Response(text, {
      status: upstream.status,
      headers: {
        "content-type": upstream.headers.get("content-type") || "application/json",
        "x-kerenscope-demo": "1",
      },
    });
  },
};
