#open -a docker

docker build . -t go-video-test:1.0.0

caffeinate docker run -v /Volumes/Tyrexl_III/docker_test:/app/videos --add-host localhost:127.0.0.1 -p 8080:8080 -it go-video-test:1.0.0 sh