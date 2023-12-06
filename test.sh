docker build -t ghcr.io/freggy/chunks76k:chunker .
docker tag ghcr.io/freggy/chunks76k:chunker reg1.chunks.76k.io/proggers/pg:1338
docker run --rm reg1.chunks.76k.io/proggers/pg:1338
