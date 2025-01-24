-- name: ListMedia :many
SELECT * FROM media
ORDER BY title, year ASC;

-- name: GetPoster :one
SELECT poster FROM media
WHERE LOWER(media.title)=LOWER(sqlc.arg(title));

-- name: LookUpMedia :many
SELECT * FROM media 
WHERE LOWER(media.title)=LOWER(sqlc.arg(title)) AND ABS(media.year-sqlc.arg(year))<=sqlc.arg(tolerance);
