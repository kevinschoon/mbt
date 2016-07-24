# marathon-bk
Simple tool to backup [Marathon](https://mesosphere.github.io/marathon/) 
application configurations via REST API. A more thorough solution is to backup 
Marathon's Zookeeper state however I have found that to be inconvient... hence
marathon-bk!

### Installation

    go get github.com/kevinschoon/marathon-bk
    

### Usage

#### Backup

    marathon-bk backup --endpoint=http://localhost:8080 --user admin:admin
    
#### Restore

    marathon-bk restore --endpoint=http://localhost:8080 --user admin:admin
