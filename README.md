pepper-server-server

# How compile .proto
- for go:
    
    In powershell:
    ```
    docker run -v ${pwd}:/defs namely/protoc-all -f service.proto -l go
    ```

    from: https://github.com/namely/docker-protoc