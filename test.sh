docker build -t ghcr.io/freggy/chunks76k:chunker .
docker run --network chunks76k_default --name test -p 5080:5080 -d -e DB_URL="postgres://chunks:chunks@chunks76k-db-1:5432/chunks?sslmode=disable" ghcr.io/freggy/chunks76k:chunker
sleep 1
curl http://localhost:5080 -d '{"type": "PUSH_ARTIFACT","occur_at": 1701565234,"operator": "test","event_data": {"resources":[{"digest": "sha256:189d0d24eedfe5a03d867f47a1a75a552d82c6a479a84903520706b3ac6c3b65","tag": "1338","resource_url": "reg1.chunks.76k.io/proggers/pg:1338"}],"repository": {"date_created": 1701495306,"name": "pg","namespace": "proggers","repo_full_name": "proggers/pg","repo_type": "private"}}}'
docker logs -f test
