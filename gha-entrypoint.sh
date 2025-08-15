#!/bin/sh -l

set -e

case "${INPUT_COMMAND}" in
    update|check) ;;
    *) echo "Error: Invalid command '${INPUT_COMMAND}'. Must be 'update' or 'check'" >&2; exit 1 ;;
esac

case "${INPUT_ALLOW}" in
    major|minor|patch) ;;
    *) echo "Error: Invalid allow value '${INPUT_ALLOW}'. Must be 'major', 'minor', or 'patch'" >&2; exit 1 ;;
esac

case "${INPUT_VERBOSE}" in
    ""|false|true) ;;
    *) echo "Error: verbose must be 'true' or 'false'" >&2; exit 1 ;;
esac

case "${INPUT_NO_SUMMARY}" in
    ""|false|true) ;;
    *) echo "Error: no-summary must be 'true' or 'false'" >&2; exit 1 ;;
esac

case "${INPUT_DRY_RUN}" in
    ""|false|true) ;;
    *) echo "Error: dry-run must be 'true' or 'false'" >&2; exit 1 ;;
esac

ARGS="${INPUT_COMMAND} --allow=${INPUT_ALLOW}"
[ "${INPUT_VERBOSE}" = "true" ] && ARGS="$ARGS --verbose"
[ -n "${INPUT_CONFIG}" ] && ARGS="$ARGS --config=${INPUT_CONFIG}"
[ "${INPUT_NO_SUMMARY}" = "true" ] && ARGS="$ARGS --no-summary"
[ "${INPUT_DRY_RUN}" = "true" ] && ARGS="$ARGS --dry-run"

exec /app/pre-commit-bump $ARGS