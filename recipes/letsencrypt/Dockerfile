FROM nginx:1.19.3

ENV LETS_ENCRYPT_DIR /etc/letsencrypt/live
ENV NGINX_CONF_DIR /etc/nginx
ENV NGINX_WRAPPER_VERSION v0.0.2
ENV NGINX_WRAPPER_CHECKSUM baad2c58df709688fed24c345e05d647c669aa025273f192f445c73869353f79

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