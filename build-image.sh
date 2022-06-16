#!/bin/sh
docker build -f Dockerfile.kong -t kong-api-gateway:latest .

while getopts 'pa:' OPTION; do
case "$OPTION" in
    a)
        export AWS_PROFILE="$OPTARG"
        echo "\n\nAws Profile Switched to : '$AWS_PROFILE'"
        ;;
    p)
        aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 602037364990.dkr.ecr.us-east-1.amazonaws.com
        docker tag kong-api-gateway:latest 602037364990.dkr.ecr.us-east-1.amazonaws.com/kong-api-gateway:latest
        docker push 602037364990.dkr.ecr.us-east-1.amazonaws.com/kong-api-gateway:latest
        ;;
    ?)
        echo "Script Usage: -p (push) -a (aws profile name)"
        ;;
esac
done