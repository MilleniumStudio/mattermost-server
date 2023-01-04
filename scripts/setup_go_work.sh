#!/bin/bash

if [[ $1 != "true" ]] ;
then
    echo "Creating a go.work file"

    txt="go 1.18\n\nuse ./\n"
    
    if [ "$BUILD_BOARDS" == "true" ] 
    then
        txt="${txt}use ../focalboard/server\nuse ../focalboard/mattermost-plugin\n"
    fi
    
    printf "$txt" > "go.work"
fi 
