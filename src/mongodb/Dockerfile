FROM mongo:latest
RUN apt-get update && apt-get install -y wget gnupg lsb-release \
    && wget -qO - https://www.mongodb.org/static/pgp/server-8.0.asc | apt-key add - \
    && echo "deb [ arch=amd64,arm64 ] http://repo.mongodb.org/apt/ubuntu $(lsb_release -cs)/mongodb-org/8.0 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-8.0.list \
    && apt-get update && apt-get install -y mongodb-mongosh
