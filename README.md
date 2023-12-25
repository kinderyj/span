Three ways to export tracing to jaeger.

1. otlp with http
2. otlp with grpc
3. opentracing

## 启动jaeger
```sh
sudo docker run -d -p 5775:5775/udp -p 16686:16686 -p 14250:14250 -p 14268:14268 -p 4317:4317 -p4318:4318 jaegertracing/all-in-one:latest
端口 4317 用于grpc
端口 4318 用于http
``` 
