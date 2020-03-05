# KittehImageBoard

## Config for s3
To use amazon s3, you will still use the file backend but you'll need to use rclone or something to mount the s3 filesystem. You can use that and then change the content and thumbnail URLs to the s3 bucket with public read permissions.
