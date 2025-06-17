-- READS --

-- name: ListMedia :many
SELECT * FROM media
ORDER BY concat(year, '-01-01') DESC, title ASC;

-- name: ListPaths :many
SELECT id, path FROM media;

-- name: GetPoster :one
SELECT poster FROM media
WHERE LOWER(media.title)=LOWER(sqlc.arg(title));

-- name: LookUpMedia :many
SELECT * FROM media 
WHERE LOWER(media.title)=LOWER(sqlc.arg(title)) AND ABS(media.year-sqlc.arg(year))<=sqlc.arg(tolerance);

-- WRITES --

-- name: InsertOrReplaceMedia :exec
INSERT OR REPLACE INTO media(id, media_type, title, year, overview, director, poster, path)
VALUES(?,?,?,?,?,?,?,?);

-- name: UpdatePath :exec
UPDATE media
SET path = ?
WHERE id = ?;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE path = ?;
