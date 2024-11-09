# apexio
Apexio is a log management and analysis platform. It aims to provide real-time insights, proactive monitoring, and comprehensive log management capabilities.


## project structure 
for ease of development and management this repository is
currently a monolithic one but in a way that it can be
seperated pretty easily, 
each directory is a service which will have its separate
mod file and dockerfile

```
.
├── LICENSE
├── README.md
├── elastic_search_cluster
│   ├── Dockerfile
│   └── main.go
├── kafka_cluster
│   ├── Dockerfile
│   └── main.go
├── log_ingestion_service
│   ├── sourcegrpc
│   └── sourceweb
│       ├── Dockerfile
│       ├── go.mod
│       ├── go.sum
│       └── main.go
├── log_processing_service
│   ├── Dockerfile
│   └── main.go
├── visualization_service
│   ├── Dockerfile
│   └── main.go
└── web-service
    ├── Dockerfile
    └── main.go
```
