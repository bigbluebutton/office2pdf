FROM debian:stretch

ENV DEBIAN_FRONTEND noninteractive

ENV REFRESHED_AT 2018-07-04 5:14PM EST

RUN apt-get update \
 && apt-get -y install libreoffice fonts-* hyphen-* unoconv

RUN mkdir /tmpl
COPY ./tmpl/upload.html /tmpl/upload.html
 
EXPOSE 8088

COPY ./office2pdf /office2pdf

CMD [ "/office2pdf" ]

