#!/bin/bash

env -i PATH="$PATH" docker compose -f docker-compose.test.yml --env-file apps/backend/.env.test "$@"
