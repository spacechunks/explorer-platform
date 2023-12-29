docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker build -t chunker .
docker run --name test -p 5080:5080 -d -e DB_URL="postgres://chunks:chunks@chunks76k-db-1:5432/chunks?sslmode=disable" -e CHUNKER_SRC_OCI_REG_USER="robot\$chunks-system" -e CHUNKER_SRC_OCI_REG_PASS="QenxGhigrVcomFWh5qaG2H3DvBWwZV7q" -e CHUNKER_DST_OCI_REG_USER="robot\$chunks-system" -e CHUNKER_DST_OCI_REG_PASS="QenxGhigrVcomFWh5qaG2H3DvBWwZV7q" -e CHUNKER_DST_OCI_REG_URL="reg1.chunks.76k.io" chunker
sleep 1
curl http://localhost:5080 -d '{"type": "PUSH_ARTIFACT","occur_at": 1701565234,"operator": "test","event_data":{"resources":[{"digest": "sha256:189d0d24eedfe5a03d867f47a1a75a552d82c6a479a84903520706b3ac6c3b65","tag": "1.0.0","resource_url": "reg1.chunks.76k.io/bergwerklags/flash:1.0.0"}],"repository": {"date_created": 1701495306,"name": "flash","namespace": "chunks-system","repo_full_name": "chunks-system/bergwerklags/flash","repo_type": "private"}}}'
docker logs -f test
