Chat Usage

    Hi
    How are You
    What is mica Temperature range 
    what is mica storage range
    how about the humidity range 
    memory and storage details of mica


Commands Usefull

    cd .. && cd PeriChat && cd cli && cd train 
    cd .. && cd train && time go run train.go -d Corpus/en -m -o ../chat/perimicaCorpustrial.gob
    cd .. && cd chat && go run chat.go -c perimicaCorpustrial.gob -t 1 -dev -anim=false
    go run train.go -d Corpus/en -m -o ../chat/PMFuncOverview.gob -config ../config_local.yaml
    go run .\chat.go -config ..\config_local.yaml
    docker build --no-cache -t perichat .
    docker run -it --rm perichat    


docker-compose run perichat
docker-compose build
docker-compose up -d 
docker-compose down
docker-compose logs perichatweb 