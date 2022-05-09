pepper-server-server

# How compile .proto
- for go:
    
    In powershell:
    ```
    docker run -v ${pwd}:/defs namely/protoc-all -f service.proto -l go
    ```

    from: https://github.com/namely/docker-protoc
# Self-sign TLS certificate
```
openssl req -newkey rsa:4096 -nodes -keyout key.pem -x509 -days 36500 -out certificate.pem
```