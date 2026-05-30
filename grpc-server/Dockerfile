FROM ruby:3.3-slim

WORKDIR /app

COPY Gemfile Gemfile.lock ./

RUN apt-get update \
 && apt-get install -y build-essential \
 && bundle config set --local without 'development test' \
 && bundle install \
 && apt-get purge -y --auto-remove build-essential \
 && rm -rf /var/lib/apt/lists/*

COPY . .

RUN adduser --disabled-password --gecos "" appuser
USER appuser

ENV PORT=8080
EXPOSE 8080

CMD ["ruby", "server.rb"]
