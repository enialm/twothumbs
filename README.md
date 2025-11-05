# Two Thumbs

Two Thumbs is a user feedback system featuring

- no fancy metrics,
- no dashboards,
- no accounts,
- no SDKs, and
- no distractions.

It provides an API to post feedback to, AI-powered digests (delivered to Slack), and not much more, by design. Two Thumbs was originally a commercial product. The entire codebase is now open source.

## API Call Example

```bash
curl -X POST 'https://your-instance.com/feedback' \
    -H 'Content-Type: application/json' \
    -H 'X-API-Key: $secret' \
    -d '{
          "prompt": "Kool?",
          "thumb_up": true,
          "comment": "& the Gang 🪩",
          "origin": "Two Thumbs",
          "category": "Landing Page",
          "in_production": true,
          "user_id": $uid
        }'
```

## Getting Started

Since Two Thumbs is rather niche, and requires initial configuration (you must, e.g., create a Slack App), I will be happy to personally assist you in getting started. Please reach out via contact@messier.ch.
