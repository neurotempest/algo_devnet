FROM docker.io/algorand/stable:3.2.1@sha256:0a87978492680fd98e2cc410f59f2bfd7fef979f035eb060685d56a6e916b5bd

RUN mkdir -p /setup
ADD algo_template.json /setup/

RUN ./goal network create -n sandnet -r /network -t /setup/algo_template.json && echo rawr

