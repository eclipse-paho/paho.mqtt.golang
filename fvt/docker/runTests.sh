#!/bin/sh

# Run the Paho Go FVT tests against a Mosquitto container.
# Arguments are passed to go test, for example:
#   ./runTests.sh -race

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

if docker compose version >/dev/null 2>&1; then
    compose() {
        docker compose "$@"
    }
elif command -v docker-compose >/dev/null 2>&1; then
    compose() {
        docker-compose "$@"
    }
else
    echo "Error: Docker Compose is not installed." >&2
    exit 127
fi

cleanup() {
    status=$?
    trap - 0 HUP INT TERM
    compose down
    exit "$status"
}

wait_for_mosquitto() {
    attempts=30

    while [ "$attempts" -gt 0 ]; do
        if compose exec -T mosquitto \
            mosquitto_pub -h 127.0.0.1 -p 1883 \
            -t paho-fvt/ready -n >/dev/null 2>&1; then
            return 0
        fi

        if [ -z "$(compose ps -q mosquitto)" ]; then
            break
        fi

        attempts=$((attempts - 1))
        sleep 1
    done

    echo "Error: Mosquitto did not become ready." >&2
    compose ps -a >&2
    compose logs --no-color mosquitto >&2
    return 1
}

trap cleanup 0 HUP INT TERM

compose up -d || exit $?
wait_for_mosquitto || exit $?

export TEST_FVT_ADDR=127.0.0.1

# --count 1 prevents Go from using cached test results. Running the tests
# repeatedly without resetting the broker may leave it in an unexpected state.
go test --count 1 -v "$@" ../../
status=$?

if [ "$status" -ne 0 ]; then
    echo "Error" >&2
fi

exit "$status"
