#!/bin/bash

# create containers (detached)

docker-compose up -d

# list containers 

docker image ls
docker container ls
docker container ls -a

# run containers

docker run my_container

# stop containers

docker stop my_container

# remove containers

docker container rm my_container

# remove all containers that are not running

docker container rm $(docker ps -a -q)

# MORE DOCKER COMMANDS: https://towardsdatascience.com/15-docker-commands-you-should-know-970ea5203421