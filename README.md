# mbt

[![CircleCI](https://img.shields.io/circleci/project/github/mesanine/mbt.svg)]()

Simple tool to backup [Marathon](https://mesosphere.github.io/marathon/) 
application configurations via REST API. 

### Installation

    go get github.com/mesanine/mbt
    

### Usage


    Usage: mbt [OPTIONS] COMMAND [arg...]

    Marathon Backup Tool

    Options:
      -e, --endpoint="http://localhost:8080"   Marathon endpoint e.g. http://localhost:8080
      -u, --user=""                            HTTP Basic Auth user:password

    Commands:
      backup       Backup the given Marathon endpoint
      restore      Restore the given Marathon endpoint

    Run 'mbt COMMAND --help' for more information on a command.

#### Backup

    mbt --endpoint=http://localhost:8080 --user admin:admin backup ./my-backup-path
    
#### Restore

    mbt --endpoint=http://localhost:8080 --user admin:admin restore ./my-restore-path


#### TODO

    * tar/gzip backup files
    * remote save / restore path
