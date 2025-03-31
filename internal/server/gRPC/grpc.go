package gRPC

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "metrics/internal/server/proto"
	"net"
	"os"
	"time"
)

type GRPCServer struct {
	auth   *auth
	Server *grpc.Server
	logger *logrus.Logger
}

type auth struct {
	cryptoKey     string
	hashKey       string
	trustedSubnet *net.IPNet
}

// Создание роутера
func NewServer(cryptoKey string, hashKey string, trustedSubnet *net.IPNet, storageCommands *StorageCommands, logger *logrus.Logger) *GRPCServer {
	instance := &GRPCServer{
		auth: &auth{
			cryptoKey:     cryptoKey,
			hashKey:       hashKey,
			trustedSubnet: trustedSubnet,
		},
		logger: logger,
	}

	interceptors := []grpc.UnaryServerInterceptor{
		instance.withLogger,
		instance.withTrustedSubnet,
		instance.withHash,
		instance.withDecrypt,
	}

	instance.Server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...))

	pb.RegisterHandlersServer(instance.Server, NewHandler(storageCommands))

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
	if g.auth.cryptoKey != "" {
		var privatePEM []byte
		// Чтение pem файла
		privatePEM, err = os.ReadFile(g.auth.cryptoKey)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}
		// Поиск блока приватного ключа
		privateKeyBlock, _ := pem.Decode(privatePEM)
		// Парсинг приватного ключа
		var privateKey *rsa.PrivateKey
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}
		if err = privateKey.Validate(); err != nil {
			return nil, fmt.Errorf("invalid private key: %v", err)
		}

		// Установка длины частей публичного ключа
		blockLen := privateKey.PublicKey.Size()

		body := req.(*pb.PostUpdatesRequest).Metrics

		// Дешифровка тела запроса частями
		var decryptedBytes []byte
		for start := 0; start < len(body); start += blockLen {
			end := start + blockLen
			if start+blockLen > len(body) {
				end = len(body)
			}

			var decryptedChunk []byte
			decryptedChunk, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, body[start:end])
			if err != nil {
				return nil, fmt.Errorf("unable to decrypt request: %v", err)
			}

			decryptedBytes = append(decryptedBytes, decryptedChunk...)
		}

		// Подмена тела запроса
		req.(*pb.PostUpdatesRequest).Metrics = decryptedBytes
	}

	return handler(ctx, req)
}
