# KittehImageBoard


## Requirements
- Imagemagick
- Go Compiler

### Install on FreeBSD
- `pkg install go imagemagick7`

## How to setup?
- Run `go build` to build the executable.
- Copy over settings_example.yaml to settings.yaml and change settings.
- Run kittehbooru and follow instructions in terminal to set up.

## Recommended way of running for scaling (100k+ posts)
- Get multiple VMs/Containers, copy over booru configs and run a instance of the program.
- Use load balancers across all instances.
- Configure shared storage, preferably using rclone to mount s3 as a filesystem
- If using s3 set the bucket to public read access and set contentURL and thumbnailURL to the http endpoint for the s3 bucket.
- Run a few instances / cluster of instances across the world, using postgresdb replication or [Amazon RDS](https://aws.amazon.com/rds/postgresql)
- CPU usage will be higher priority as generating thumbnails and searching requires a lot of CPU, you may need more ram for more users for caches.


## What is this for?
- It's primarily a replacement for Gelbooru/Danbooru / other tagged imageboard solutions.
- It's PHP-free and designed to be high-performance and scalable.
