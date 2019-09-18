#!/bin/bash

function add_cluster() {
    [ -z "$1" ] && usage
    [ -z "$2" ] && usage
    curl -i -XPOST http://127.0.0.1:8000/api/v1/cluster -d "{\"name\": \"$1\", \"prefix\": \"$2\"}"
}

function remove_cluster() {
    [ -z "$1" ] && usage
    curl -i -XDELETE http://127.0.0.1:8000/api/v1/cluster -d "{\"name\": \"$1\"}"
}

function add_endpoint() {
    [ -z "$1" ] && usage
    [ -z "$2" ] && usage
    [ -z "$3" ] && usage
    curl -i -XPOST http://127.0.0.1:8000/api/v1/endpoint -d "{\"cluster\": \"$1\", \"ipaddress\": \"$2\", \"port\": $3}"
}

function remove_endpoint() {
    [ -z "$1" ] && usage
    [ -z "$2" ] && usage
    [ -z "$3" ] && usage
    curl -i -XDELETE http://127.0.0.1:8000/api/v1/endpoint -d "{\"cluster\": \"$1\", \"ipaddress\": \"$2\", \"port\": $3}"
}

function commit_changes() {
    curl -i -XPOST http://127.0.0.1:8000/api/v1/commit
}

function usage() {
    echo "Usage:"
    echo "$0 cluster add <name> <url_prefix>"
    echo "$0 cluster remove <name>"
    echo "$0 endpoint add <cluster_name> <ip_address> <port>"
    echo "$0 endpoint remove <cluster_name> <ip_address> <port>"
    echo "$0 commit"
    exit 1
}

case "$1" in
    cluster)
        if [ "$2" == "add" ]; then
            add_cluster "$3" "$4"
        elif [ "$2" == "remove" ]; then
            remove_cluster "$3"
        else
            usage
        fi
      ;;
  endpoint)
        if [ "$2" == "add" ]; then
            add_endpoint "$3" "$4" "$5"
        elif [ "$2" == "remove" ]; then
            remove_endpoint "$3" "$4" "$5"
        else
            usage
        fi
      ;;

    commit)
        commit_changes
        ;;
    *)
        usage
        ;;
esac
