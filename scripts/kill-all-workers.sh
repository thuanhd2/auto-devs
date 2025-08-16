#! /bin/bash

PIDS_FOLDER="/private/var/folders/tv/531lt6yx3ss28h1b7bcpb1900000gn/T/autodevs"

# read all files in the PIDS_FOLDER
for pid_file in "$PIDS_FOLDER"/*.pid; do
    file_name=$(basename "$pid_file")
    echo "Killing process $file_name"
    pid=$(cat "$pid_file")
    echo "PID: $pid"
    # kill the process
    kill "$pid"
    # remove the file
    rm "$pid_file"
done