package gRPC

import (
	"context"
	"metrics/internal/server/proto"

	"crypto/x509"
	"encoding/pem"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"time"
)

type GRPCServer struct {
	cryptoKey string
	Server    *grpc.Server
}

func NewServer(cryptoKey string) *GRPCServer {
	instance := &GRPCServer{
		cryptoKey: cryptoKey,
	}
	interceptors := []grpc.UnaryServerInterceptor{
		instance.withLogger,
		instance.withTrustedSubnet,
		instance.withHash,
	}

	if cryptoKey != "" {
		interceptors = append(interceptors, instance.withDecrypt)
	}

	instance.Server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...))

	return instance
}

func (g *GRPCServer) withLogger(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// выполняем действия перед вызовом метода
	start := time.Now()

	fmt.Println("METRICS", req)

	// вызываем RPC-метод
	resp, err = handler(ctx, req)

	fmt.Println("FINISHED REQUEST", req, time.Since(start))

	return resp, err
}

func (g *GRPCServer) withTrustedSubnet(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

	fmt.Println("SUBNET MIDDLEWAER", req)
	return handler(ctx, req)
}

func (g *GRPCServer) withHash(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return handler(ctx, req)
}

func (g *GRPCServer) withDecrypt(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// Чтение pem файла
	privatePEM, err := os.ReadFile(g.cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}
	// Поиск блока приватного ключа
	privateKeyBlock, _ := pem.Decode(privatePEM)
	// Парсинг приватного ключа
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return
	}

	// Установка длины частей публичного ключа
	blockLen := privateKey.PublicKey.Size()

	fmt.Println("BLOCK LEN", blockLen)

	fmt.Println("REQUEST", req.(*proto.PostUpdatesRequest).Metrics)

	//// Дешифровка тела запроса частями
	//var decryptedBytes []byte
	//for start := 0; start < len(req); start += blockLen {
	//	end := start + blockLen
	//	if start+blockLen > len(body) {
	//		end = len(body)
	//	}
	//
	//	var decryptedChunk []byte
	//	decryptedChunk, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, body[start:end])
	//	if err != nil {
	//		s.logger.Errorf("error decrypting random text: %s", err.Error())
	//		w.WriteHeader(http.StatusBadRequest)
	//		return
	//	}
	//
	//	decryptedBytes = append(decryptedBytes, decryptedChunk...)
	//}
	//
	//// Подмена тела запроса
	//r.Body = io.NopCloser(bytes.NewBuffer(decryptedBytes))

	return handler(ctx, req)
}
