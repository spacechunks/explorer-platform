flux push artifact oci://ghcr.io/freggy/chunks76k:latest --path=wlc --source=git@github.com:Freggy/chunks76k.git --revision="$(git rev-parse HEAD)"
