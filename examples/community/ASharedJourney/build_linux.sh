#!/bin/bash

DESTINATION="/tmp/ASharedJourney"

make build_assets
go build main.go
chmod +x main
mkdir -p $DESTINATION
cp main $DESTINATION/ASharedJourney
cp -r ./assets $DESTINATION/

echo "Built!"
echo "You may now send" $DESTINATION "to your beautiful friends!"
