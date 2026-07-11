# Bakes the seed dataset into the goccy BigQuery emulator image. RunContainer has
# no volume-mount option and the emulator image is distroless, so the seed is
# copied in at build time and loaded with --data-from-yaml at run time.
FROM ghcr.io/goccy/bigquery-emulator:latest
COPY seed.yaml /seed.yaml
