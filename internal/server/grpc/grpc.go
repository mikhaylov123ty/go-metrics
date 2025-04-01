package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"net"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "metrics/internal/server/proto"
	"metrics/internal/server/utils"
)

// GRPCServer - структура инстанса gRPC сервера
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

// NewServer создает инстанс gRPC сервера
func NewServer(cryptoKey string, hashKey string, trustedSubnet *net.IPNet, storageCommands *StorageCommands, logger *logrus.Logger) *GRPCServer {
	instance := &GRPCServer{
		auth: &auth{
			cryptoKey:     cryptoKey,
			hashKey:       hashKey,
			trustedSubnet: trustedSubnet,
		},
		logger: logger,
	}

	// Определение перехватчиков
	interceptors := []grpc.UnaryServerInterceptor{
		instance.withLogger,
		instance.withTrustedSubnet,
		instance.withHash,
		instance.withDecrypt,
	}

	//Регистрация инстанса gRPC с перехватчиками
	instance.Server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...))

	pb.RegisterHandlersServer(instance.Server, NewHandler(storageCommands))

	return instance
}

// withLogger - перехватчик логирует запросы
func (g *GRPCServer) withLogger(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// Запуск таймера
	start := time.Now()

	g.logger.Infof("gRPC server received request: %s", info.FullMethod)

	// Запуск RPC-метода
	resp, err = handler(ctx, req)

	// Логирует код и таймер
	e, _ := status.FromError(err)
	g.logger.Infof("Request completed with code %v in %s", e.Code(), time.Since(start))

	return resp, err
}

// withTrustedSubnet - перехватчик проверяет подсеть в метаданных
func (g *GRPCServer) withTrustedSubnet(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// Проверка наличия записи подсети
	if g.auth.trustedSubnet != nil {
		g.logger.Infof("start checking request subNet")

		// Чтение метаданных
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "can't extract metadata from request")
		}
		header, ok := meta["x-real-ip"]
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "can't extract hash header from request")
		}
		requestIP := net.ParseIP(header[0])
		if requestIP == nil {
			return nil, status.Errorf(codes.InvalidArgument, "error parsing X-Real-IP header")
		}

		// Проверка
		if !g.auth.trustedSubnet.Contains(requestIP) {
			return nil, status.Errorf(codes.PermissionDenied, "IP address is not trusted")
		}
	}

	return handler(ctx, req)
}

// withHash - перехватчик проверяет наличие хеша в метаданных и сверяет с телом запроса
func (g *GRPCServer) withHash(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// Проверка наличия флага ключа
	if g.auth.hashKey != "" {
		g.logger.Infof("start checking gRPC request hash")

		// Чтеные метаданных
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "can't extract metadata from request")
		}
		var requestHeader []byte
		header, ok := meta["hashsha256"]
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "can't extract hash header from request")
		}
		requestHeader, err = hex.DecodeString(header[0])
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "can't decode hash header from request")
		}

		// Чтение тела запроса
		body := req.(*pb.PostUpdatesRequest).Metrics

		// Вычисление и валидация хеша
		hash := utils.GetHash(g.auth.hashKey, body)
		if !hmac.Equal(hash, requestHeader) {
			return nil, status.Errorf(codes.PermissionDenied, "hash does not match")
		}
	}

	return handler(ctx, req)
}

// withDecrypt - перехватчик дешифровки тела запроса
func (g *GRPCServer) withDecrypt(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// Проверка наличия флага приватного ключа
	if g.auth.cryptoKey != "" {
		g.logger.Infof("start decrypt gRPC request")

		// Чтение pem файла
		var privatePEM []byte
		privatePEM, err = os.ReadFile(g.auth.cryptoKey)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "unable to read private key: %v", err)
		}

		// Поиск блока приватного ключа
		privateKeyBlock, _ := pem.Decode(privatePEM)

		// Парсинг приватного ключа
		var privateKey *rsa.PrivateKey
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unable to parse private key: %v", err)
		}
		if err = privateKey.Validate(); err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid private key: %v", err)
		}

		// Установка длины частей публичного ключа
		blockLen := privateKey.PublicKey.Size()

		// Чтение метрик
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
				return nil, status.Errorf(codes.Internal, "unable to decrypt request: %v", err)
			}

			decryptedBytes = append(decryptedBytes, decryptedChunk...)
		}

		// Подмена тела запроса
		req.(*pb.PostUpdatesRequest).Metrics = decryptedBytes
	}

	return handler(ctx, req)
}
