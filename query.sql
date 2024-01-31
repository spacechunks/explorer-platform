-- name: ListVariantDeploys :many
SELECT * FROM variant_deployments WHERE mode = $1;
