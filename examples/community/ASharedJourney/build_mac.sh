#!/bin/bash

SOURCE="./build/Template.app"
DESTINATION="./build/ASharedJourney.app"

make build_assets
go build main.go
chmod +x main
rm -rf $DESTINATION
cp -r $SOURCE $DESTINATION
cp main $DESTINATION/Contents/MacOS/ASharedJourney

echo "Built!"
echo "You may now send" $DESTINATION "to your beautiful friends!"
