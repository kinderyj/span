package main

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// 计算斐波那契队列
func Fibonacci(n uint) (uint64, error) {
	if n <= 1 {
		return uint64(n), nil
	}

	var n2, n1 uint64 = 0, 1
	for i := uint(2); i < n; i++ {
		n2, n1 = n1, n1+n2
	}

	return n2 + n1, nil
}

// 这里模拟服务之间的掉用的效果.
func Run(ctx context.Context) {
	_, span := otel.Tracer("hello1").Start(ctx, "two")
	defer span.End()
	fibonacci, err := Fibonacci(20)
	if err != nil {
		return
	}
	fmt.Println(fibonacci)
}

// newExporter returns an OTLP exporter.
func newExporter(url string) (*otlptrace.Exporter, error) {
	return otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithInsecure(),
	)
}

// 资源是一种特殊类型的属性，适用于进程生成的所有跨度。这些应该用于表示有关非临时进程的底层元数据
// 例如，进程的主机名或其实例 ID
func newResource() (*resource.Resource, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fibqqqq"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// 主函数
func main() {
	url := "http://127.0.0.1:4317"
	os.Setenv("OTEL_SERVICE_NAME", "t2")
	// 创建导出器
	exp, err := newExporter(url)
	if err != nil {
		fmt.Println("Error creating exporter:", err)
		return
	}

	// 创建链路生成器, 这里将导出器与资源信息配置进去.
	resource, err := newResource()
	if err != nil {
		fmt.Println("Error creating resource:", err)
		return
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource),
	)

	// 结束时关闭链路生成器
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Println(err)
		}
	}()

	// 将创建的生成器设置为全局变量.
	// 这样直接使用otel.Tracer就可以创建链路.
	// 否则 就要使用 tp.Tracer的形式创建链路.
	otel.SetTracerProvider(tp)
	newCtx, span := otel.Tracer("hello").Start(context.Background(), "one")
	Run(newCtx)
	span.End()
}
