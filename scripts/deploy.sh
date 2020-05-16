#!/bin/bash

# this is straightforward deploy to remote that does the job for me for now.
# probably later I'll need to add a hook to deploy on merge

# how to use:
# 1. copy this file to ../bin or elsewhere you like (../bin is convenient because it is in project dir and out of git)
# 2. set variables that is just right below
# 3. run this script

# remote's requirements:
# 1. supervisor must be installed and config for supervisor to run server must be created
#     (look at ../scripts/supervisor.conf for example)

# set variables required for deploy

# server's host & port
serverHost=
serverPort=
# local project's dir to deploy from
projectDir=
# remote project's dir to deploy to
remoteProjectDir=
# remote's user
remoteUser=
# remote's addr
remoreAddr=
# database conf
dbAddr=localhost:5432
dbName=postgres
dbUser=postgres
dbPassword=postgres

# necessary checks
if [ -z "$remoreAddr"  ]
then
  echo "Unable to deploy: Remote addr is not set"
  exit
fi
if [ -z "$remoteUser"  ]
then
  echo "Unable to deploy: Remote user is not set"
  exit
fi
if [ -z "$dbAddr"  ]
then
  echo "Unable to deploy: Database addr is not set"
  exit
fi
if [ -z "$dbName"  ]
then
  echo "Unable to deploy: Database name is not set"
  exit
fi
if [ -z "$dbUser"  ]
then
  echo "Unable to deploy: Database user is not set"
  exit
fi
if [ -z "$dbPassword"  ]
then
  echo "Unable to deploy: Database password is not set"
  exit
fi
if [ -z "$serverHost"  ]
then
  echo "Unable to deploy: Server's host is not set"
  exit
fi
cd $projectDir || exit

# test api
cd ./server/test || exit
echo "Testing API..."
go test || { echo "Unable to deploy: tests failed"; exit; }

echo "Compiling API server..."
cd $projectDir || exit
go build -ldflags "-s -w" -o ./bin/refto-server cmd/server/main.go || exit

echo "Compiling CLI..."
go build -ldflags "-s -w" -o ./bin/refto-cli cmd/cli/main.go || exit

echo "Copying binaries to remote (${remoteUser}@${remoreAddr})..."
scp ./bin/refto-server ./bin/refto-cli ${remoteUser}@${remoreAddr}:~/

echo "Setting up server on remote..."
ssh -T ${remoteUser}@${remoreAddr} << EOF
echo "Stopping supervisor..."
sudo service supervisor stop || exit

echo "Moving binaries..."
mv ~/refto-server $remoteProjectDir/server || exit
mv ~/refto-cli $remoteProjectDir/cli || exit

echo "Writing config..."
/bin/cat <<EOM > $remoteProjectDir/.config.yaml
app_env: dev

db:
  addr: $dbAddr
  user: $dbUser
  password: $dbPassword
  name: $dbName
  log_queries: true

server:
  host: $serverHost
  port: $serverPort
  api_base_path: api

dir:
  data: "$remoteProjectDir/data"
  logs: ""
EOM

echo "Migrating database..."
cd $remoteProjectDir || exit
./cli migrate || exit

echo "Starting supervisor..."
sudo service supervisor start || exit
EOF

echo "Server deployed!"