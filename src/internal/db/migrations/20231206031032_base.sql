-- migrate:up
CREATE TABLE variant_deployments(
    project VARCHAR(100),
    mode VARCHAR(100),
    variant VARCHAR(100),
    cluster_url VARCHAR(100),
    cluster_token VARCHAR(100)
);

-- migrate:down

