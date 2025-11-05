-- Note: add indexes as appropriate

CREATE TABLE installations (
    slack_workspace TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    bot_token TEXT NOT NULL
);

CREATE TABLE accounts (
    account_id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    account_expiry_date TIMESTAMPTZ NOT NULL,
    activation_code TEXT UNIQUE NOT NULL,
    api_key TEXT UNIQUE NOT NULL,
    slack_workspace TEXT,
    slack_channel TEXT,
    feedback_count INT NOT NULL DEFAULT 0
);

CREATE TABLE prompts (
    id BIGSERIAL PRIMARY KEY,
    slack_workspace TEXT NOT NULL,
    origin TEXT NOT NULL,
    category TEXT NOT NULL,
    prompt TEXT NOT NULL,
    CONSTRAINT unique_prompt UNIQUE (slack_workspace, origin, category, prompt)
);

CREATE TABLE feedback (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    slack_workspace TEXT NOT NULL,
    prompt TEXT NOT NULL,
    thumb_up BOOLEAN NOT NULL,
    comment TEXT,
    origin TEXT NOT NULL,
    category TEXT NOT NULL,
    in_production BOOLEAN NOT NULL,
    user_id TEXT NOT NULL
);

CREATE TABLE summaries (
    id BIGSERIAL PRIMARY KEY,
    summary_date DATE NOT NULL,
    slack_workspace TEXT NOT NULL,
    origin TEXT NOT NULL,
    category TEXT NOT NULL,
    prompt TEXT NOT NULL,
    n_comments INT NOT NULL,
    summary TEXT NOT NULL
);

CREATE TABLE issues (
    slack_workspace TEXT,
    origin TEXT,
    report TEXT NOT NULL,
    PRIMARY KEY (slack_workspace, origin)
);
