#open -a docker

source .env
docker build . -t go-video-test:1.0.0
echo "Running Docker Container with volume: $DOCKER_VOLUME_TEST"
caffeinate docker run -v $DOCKER_VOLUME_TEST:/app/videos --add-host localhost:127.0.0.1 -p 8080:8080 -it go-video-test:1.0.0 sh