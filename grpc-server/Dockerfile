# grpc-server/Dockerfile
FROM ruby:3.3-slim

WORKDIR /app

RUN apt-get update \
 && apt-get install -y build-essential \
 && rm -rf /var/lib/apt/lists/*

COPY Gemfile Gemfile.lock ./
RUN bundle install --without development test

COPY . .

ENV PORT=8080
EXPOSE 8080

CMD ["ruby", "server.rb"]
