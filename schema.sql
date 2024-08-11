-- Table for channel posts with food pictures. Has `raw` jsonb column for storing the raw data from Telegram, and other columns for storing the parsed data.
CREATE TABLE food_posts (
    message_id BIGINT PRIMARY KEY,
    post_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    photo_key TEXT NOT NULL,
    caption_text TEXT NOT NULL,
    raw JSONB NOT NULL
);
