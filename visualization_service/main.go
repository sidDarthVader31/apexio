package main


// order of deployment -
// 1. config maps
// 2. secrets
// 3. kafka 
// 4. create topics in kafka
// 5. elasticsearch 
// 6. create index with our mapping (api call)
// 7. source_web
// 8. source_grpc
// 9. log_processing_service
// 10. grafana
// 11. create data source grafana 
// 12. create dashboad grafana 

