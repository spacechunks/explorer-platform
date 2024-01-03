# lbwl-flash

Rewrite of the original Flash minigame in Kotlin. It has been kept as simple as possible on purpose. No configuration, no nothing.

## Contributing

If you want to add new features feel free to do so. Just comply with the standard Kotlin codestyle.

### Dev setup [Docker]

- build the project `gradle build`
- Make sure [Docker Compose](https://docs.docker.com/compose/) is installed
- navigate to `./flash_docker` and execute `docker compose build`
- execute in the same directory `docker compose up -d` to start a local container
- connect to `localhost` in minecraft

after you done some changes you only need to build the plugin and restart docker:
- in the root directory `maven package`
- `./flash_docker` execute `docker compose down` &`docker compose up -d`

## Pushing Flash to Dockerhub

Push a minecraft server with flash to the Dockerhub:

1. add the secrets `secrets.DOCKERHUB_USERNAME` and `secrets.DOCKERHUB_PASSWORD` with a dockerhub account where the image should be available
2. create a release + tag in GitHub. This will trigger an action that will build and push an image while the tag the version of the image is 
