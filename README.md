# Entain Test Task – Golang + PostgreSQL

## Overview

The test application for processing the incoming requests from the 3d-party providers.

The application have an HTTP URL to receive incoming `POST` requests.
To receive the incoming POST requests the application have an HTTP URL endpoint.
Technologies: Golang + Postgres.

## Quick Start

### Prerequisites
- Docker & Docker Compose installed

### Run the service

#### Run with Makefile
make run

#### How to run manually:
docker compose up -d --build

#### Stop the application:
docker compose down
 