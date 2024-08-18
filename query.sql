-- name: InsertFoodPost :execrows
INSERT INTO food_posts (
    message_id, post_ts, photo_key, caption_text, raw
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (message_id) DO NOTHING
RETURNING *;

-- name: FindFoodPostByMessageId :one
SELECT * FROM food_posts WHERE message_id = $1;

-- name: UpdateFoodPostSetPhoto :execrows
UPDATE food_posts SET photo_key = $2 WHERE message_id = $1;

-- name: FindFoodPostsByDateRange :many
SELECT * FROM food_posts
WHERE post_ts >= @from_ts AND post_ts <= @until_ts ORDER BY post_ts ASC;
