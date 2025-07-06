#!/bin/bash

docker compose --env-file ./apps/backend/.env "$@"
