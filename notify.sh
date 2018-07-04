#!/bin/bash
#

function docker_api {
    local scheme
    local curl_opts=(-s)
    local method=${2:-GET}
    # data to POST
    if [[ -n "${3:-}" ]]; then
        curl_opts+=(-d "$3")
    fi
    if [[ -z "$DOCKER_HOST" ]];then
        echo "Error DOCKER_HOST variable not set" >&2
        return 1
    fi
    if [[ $DOCKER_HOST == unix://* ]]; then
        curl_opts+=(--unix-socket ${DOCKER_HOST#unix://})
        scheme='http://localhost'
    else
        scheme="http://${DOCKER_HOST#*://}"
    fi
    [[ $method = "POST" ]] && curl_opts+=(-H 'Content-Type: application/json')
    curl "${curl_opts[@]}" -X${method} ${scheme}$1
}

function docker_kill {
    local id="${1?missing id}"
    local signal="${2?missing signal}"
    docker_api "/containers/$id/kill?signal=$signal" "POST"
}

function reload_nginx {
    local _docker_gen_container=$(get_docker_gen_container)
    local _nginx_proxy_container=$(get_nginx_proxy_container)

    if [[ -n "${_docker_gen_container:-}" ]]; then
        # Using docker-gen and nginx in separate container
        echo "Reloading nginx docker-gen (using separate container ${_docker_gen_container})..."
        docker_kill "${_docker_gen_container}" SIGHUP

        if [[ -n "${_nginx_proxy_container:-}" ]]; then
            # Reloading nginx in case only certificates had been renewed
            echo "Reloading nginx (using separate container ${_nginx_proxy_container})..."
            docker_kill "${_nginx_proxy_container}" SIGHUP
        fi
    fi
}

function labeled_cid {
    docker_api "/containers/json" | jq -r '.[] | select(.Labels["'$1'"])|.Id'
}

function get_docker_gen_container {
    # First try to get the docker-gen container ID from the container label.
    local docker_gen_cid="$(labeled_cid nginx_proxy_companion.docker_gen)"

    # If a container ID was found, output it. The function will return 1 otherwise.
    [[ -n "$docker_gen_cid" ]] && echo "$docker_gen_cid"
}

function get_nginx_proxy_container {
    local volumes_from
    # First try to get the nginx container ID from the container label.
    local nginx_cid="$(labeled_cid nginx_proxy_companion.nginx_proxy)"

    # If a container ID was found, output it. The function will return 1 otherwise.
    [[ -n "$nginx_cid" ]] && echo "$nginx_cid"
}

CERT_DIR=${CERT_DIR:=/certs/}

/docker-ssl-nginx-proxy-companion -server=${CERT_SERVER} -certs=${CERT_DIR}

modifiedFiles=$(find ${CERT_DIR} -mmin -1 -type f -print | wc -l)
if [[ "${modifiedFiles}" -ne "0" ]]; then

  reload_nginx 
  
fi

