# Captions validator

## Run manually step by step
### Host the mockserver
python3 server.py
### Build Docker image
docker build -t caption-validator .
### Run srt validation
docker run --rm -v "$(pwd)/samples:/data" caption-validator --t_start=0 --t_end=10 --coverage=50 --endpoint=http://host.docker.internal:8080  /data/examples.srt
## or run via shell script
chmod 755 run.sh
./run.sh 
