# viewer-config-service
De viewer-config-service serveert de configuratie gedefinieerd in [viewer-config-service-data](http://github.so.kadaster.nl/PDOK/viewer-config-service-data) voor de [pdok-viewer-angular](http://github.so.kadaster.nl/PDOK/pdok-viewer-angular).

## Docker

```sh
docker build -t pdok/viewer-config-service .
docker run -d -p 80:80 --name viewer-config-service pdok/viewer-config-service
docker stop viewer-config-service && docker rm viewer-config-service
```

## CLI

```sh
go run main.go watcher.go -c "example/config.yaml"
```

## Config

We need a config like: 

```yaml
baseurl: "https://service.pdok.nl/"
jsonfiles: "/srv/data/viewer-config-service/"
```
