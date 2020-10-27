FROM nginx:1.19.3

ENV LETS_ENCRYPT_DIR /etc/letsencrypt/live
ENV NGINX_CONF_DIR /etc/nginx
ENV NGINX_WRAPPER_VERSION v0.0.3
ENV NGINX_WRAPPER_CHECKSUM 02417ed2d8133200121a505bde3e86c6e3fb7b902f9225b07ed68acd082d34cb

COPY opt /opt

RUN set -eux \
    export DEBIAN_FRONTEND=noninteractive; \
    rm /etc/machine-id; \
    curl --retry 6 -Ls -o "${NGINX_CONF_DIR}/options-ssl-nginx.conf" \
        "https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf"; \
    curl --retry 6 -Ls -o "${NGINX_CONF_DIR}/ssl-dhparams.pem" \
        "https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem"; \
    mkdir -p /opt/nginx-wrapper/bin /opt/nginx-wrapper/plugins /opt/nginx-wrapper/run; \
    curl -o /opt/nginx-wrapper/bin/nginx-wrapper.gz --retry 6 -Ls "https://github.com/nginxinc/nginx-wrapper/releases/download/${NGINX_WRAPPER_VERSION}/nginx-wrapper-linux_amd64-${NGINX_WRAPPER_VERSION}.gz"; \
    echo "${NGINX_WRAPPER_CHECKSUM}  /opt/nginx-wrapper/bin/nginx-wrapper.gz" | sha256sum -c; \
    gunzip /opt/nginx-wrapper/bin/nginx-wrapper.gz; \
    chmod +x /opt/nginx-wrapper/bin/*; \
    apt-get update -qq; \
    apt-get install -y -qq uuid certbot python3-certbot-nginx; \
    rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/*

WORKDIR /opt/nginx-wrapper

CMD [ "bin/launch_wrapper.sh" ]

EXPOSE 80
EXPOSE 443