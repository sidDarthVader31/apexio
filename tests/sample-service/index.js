const express = require('express');
const app = express()
const { faker } = require('@faker-js/faker')
const axios = require('axios')

const requestMethod = ['POST','GET','PUT','DELETE']
const logLevel = ['INFO','ERROR','WARN','FATAL','DEBUG','UNSPECIFIED']
const environment = ['production', 'dev','qa', 'release' ,'perf', 'staging']
  const data = {
    "id": Math.random(),
    "metadata": {
      "requestId": "req-abc-123",
      "clientIp": "192.168.1.100", 
      "requestMethod":  "POST",
      "requestPath": "/api/v1/users/login", 
      "responseStatus": 200,
      "responseDuration": 156.7
    },
    "timestamp": Date.now(),
    "logLevel": "INFO",
    "message": "User login successful", "source": {
      "host": "api-server-01",
      "service": "auth-service",
      "environment": "production"
    }
  }
 const generateResponseStatus = () => {
    const statusGroups = [
      [200, 201, 204, 209],   // 2xx
      [301, 302, 304],   // 3xx
      [400, 401, 403, 404], // 4xx
      [500, 502, 503, 504] // 5xx
    ];
    const group = statusGroups[Math.floor(Math.random() * statusGroups.length)];
    return group[Math.floor(Math.random() * group.length)];
  }
const generateRandomByArray = (value) =>{
  return value[Math.floor(Math.random()* value.length)]
}

const services ={
      auth: [
        '/login', 
        '/logout', 
        '/register', 
        '/password/reset', 
        '/verify-email'
      ],
      users: [
        '/users', 
        '/users/profile', 
        '/users/settings', 
        '/users/{userId}', 
        '/users/search'
      ],
      products: [
        '/products', 
        '/products/{productId}', 
        '/products/categories', 
        '/products/search', 
        '/products/recommendations'
      ],
      orders: [
        '/orders', 
        '/orders/{orderId}', 
        '/orders/create', 
        '/orders/history', 
        '/orders/tracking'
      ],
      payments: [
        '/payments', 
        '/payments/methods', 
        '/payments/{paymentId}', 
        '/checkout', 
        '/payments/refund'
      ]
    };

const generateRandomService = () =>{
const serviceKeys = Object.keys(services);
    const randomService = serviceKeys[Math.floor(Math.random() * serviceKeys.length)];
    const endpoints = services[randomService];
    let endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  endpoint = endpoint.replace('{userId}', generateId());
    endpoint = endpoint.replace('{productId}', generateId());
    endpoint = endpoint.replace('{orderId}', generateId());
    endpoint = endpoint.replace('{paymentId}', generateId());

    return {
      service: randomService,
      path: endpoint
    };
  }


const  generateId = () => {
    return Math.random().toString(36).substring(2, 10);
  }
const getSamplePayload = () =>{
  const status  = generateResponseStatus()
   const currentTime = Date.now();
  const twentyFourHoursAgo = currentTime - (24 * 60 * 60 * 1000);
  const randomTimestamp = twentyFourHoursAgo + Math.random() * (24 * 60 * 60 * 1000);
  const {service, path} = generateRandomService();
  return {
    "id" : Math.floor(Math.random() * 100000),
    "metadata": {
      "requestId": faker.string.uuid(),
      "clientIp": faker.internet.ipv4(), 
      "userAgent": faker.internet.userAgent(),
      "requestMethod": generateRandomByArray(requestMethod),
      "requestPath": path , 
      "responseStatus": status,
      "responseDuration": Math.floor(Math.random()* 1000) + 1  , // ms
      "extra":{
         traceId: faker.string.uuid()
      }
    },
    "timestamp": Math.floor(randomTimestamp),
    "logLevel": generateRandomByArray(logLevel),
    "message": `${status} ${faker.music.artist()}`,
    "source": {
      "host": faker.internet.domainName(),
      "service": service,
      "environment": generateRandomByArray(environment)
    }
  } 
}
app.get('/generate/sample-logs/:count', async (req, res) =>{
  const count = req.params.count;
  try{
    for(let i = 0; i< count;i++){
    const data =  getSamplePayload();
    const response = await axios.default.post('http://localhost:8080/api/v1/log', data)
    console.log(`response for iteration:${i}: ${response.status}`);
    }
    return res.send(data).status(200);
  }
    catch(e){
    console.log(`error:`, e)
    return res.send(e.toString()).status(500)
  }
});
app.listen(3001, () =>{
  console.log(`sample server running on 3001`)
})
