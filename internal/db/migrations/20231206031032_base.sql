-- migrate:up
CREATE TABLE flavor_deployments(
    project VARCHAR(100),
    mode VARCHAR(100),
    flavor VARCHAR(100),
    cluster_url VARCHAR(100),
    cluster_token VARCHAR(100)
);

-- migrate:down

