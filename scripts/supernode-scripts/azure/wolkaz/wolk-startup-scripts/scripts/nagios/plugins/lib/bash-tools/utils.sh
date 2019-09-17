#!/usr/bin/env bash
#  vim:ts=4:sts=4:sw=4:et
#
#  Author: Hari Sekhon
#  Date: 2015-05-25 01:38:24 +0100 (Mon, 25 May 2015)
#
#  https://github.com/harisekhon/pytools
#
#  License: see accompanying Hari Sekhon LICENSE file
#
#  If you're using my code you're welcome to connect with me on LinkedIn and optionally send me feedback to help improve or steer this or other code I publish
#
#  https://www.linkedin.com/in/harisekhon
#

set -eu
[ -n "${DEBUG:-}" ] && set -x
srcdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export TRAP_SIGNALS="INT QUIT TRAP ABRT TERM EXIT"

hr(){
    echo "================================================================================"
}

section(){
    hr
    "$srcdir/center80.sh" "$@"
    hr
    echo
}

# TODO:
#export SPARK_HOME="$(ls -d tests/spark-*-bin-hadoop* | head -n 1)"

type isExcluded &>/dev/null || . "$srcdir/excluded.sh"

is_linux(){
    if [ "$(uname -s)" = "Linux" ]; then
        return 0
    else
        return 1
    fi
}

is_mac(){
    if [ "$(uname -s)" = "Darwin" ]; then
        return 0
    else
        return 1
    fi
}

is_jenkins(){
    if [ -n "${JENKINS_URL:-}" ]; then
        return 0
    else
        return 1
    fi
}

is_travis(){
    if [ -n "${TRAVIS:-}" ]; then
        return 0
    else
        return 1
    fi
}

is_CI(){
    if is_jenkins || is_travis; then
        return 0
    else
        return 1
    fi
}

if is_travis; then
    #export DOCKER_HOST="${DOCKER_HOST:-localhost}"
    export HOST="${HOST:-localhost}"
fi

if is_travis; then
    sudo=sudo
else
    sudo=""
fi

# useful for cutting down on number of noisy docker tests which take a long time but more importantly
# cause the CI builds to fail with job logs > 4MB
ci_sample(){
    local versions="$@"
    if [ -n "${SAMPLE:-}" ] || is_CI; then
        if [ -n "$versions" ]; then
            local a
            IFS=' ' read -r -a a <<< "$versions"
            local highest_index="${#a[@]}"
            local random_index="$(($RANDOM % $highest_index))"
            echo "${a[$random_index]}"
            return 1
        else
            if [ "$(($RANDOM % 4))" != 0 ]; then
                return 1
            fi
        fi
    else
        if [ -n "$versions" ]; then
            echo "$versions"
        fi
    fi
    return 0
}

untrap(){
    trap - $TRAP_SIGNALS
}
