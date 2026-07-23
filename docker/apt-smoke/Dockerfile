# Smoke test: install bareai from Cloudsmith APT on Debian.
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl gnupg \
    && curl -1sLf 'https://dl.cloudsmith.io/public/baselhusam/bareai/setup.deb.sh' | bash \
    && apt-get update \
    && apt-get install -y bareai \
    && bareai version \
    && apt-get purge -y curl gnupg \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

CMD ["bareai", "version"]
