package config

import (
	"errors"
	"flag"
	"net"
	"strconv"
	"strings"
	"time"
)

// NetAddress holds structured network address data for host and port.
// It implements the flag.Value interface.
type NetAddress struct {
	Host string
	Port int
}

// ParseFlags parses all configuration flags.
//
// Flags:
//
//	-a server address in format [host]:[port]
//	-grpc-address grpc server address in format [host]:[port]
//	-f file storage path
//	-d database DSN
//	-crypto-key private key path
//	-c/-config json file path with configs
//	-password-hash-key password hash key
//	-token-sign-key token signing key
//	-token-issuer token issuer name
//	-token-duration token duration (e.g., "1h", "30m")
//	-request-timeout request timeout (e.g., "30s", "1m")
//	-hash-key security hash key
func ParseFlags() *StructuredConfig {
	var serverAddress, grpcServerAddress NetAddress
	var fileStoragePath string
	var databaseDSN string
	var cryptoKey string
	var jsonConfigPath string
	var passwordHashKey string
	var tokenSignKey string
	var tokenIssuer string
	var tokenDuration time.Duration
	var requestTimeout time.Duration
	var hashKey string

	flag.Var(&serverAddress, "a", "Net address host:port")
	flag.Var(&grpcServerAddress, "grpc-address", "Net grpc server address host:port")
	flag.StringVar(&fileStoragePath, "f", "", "File storage path")
	flag.StringVar(&databaseDSN, "d", "", "Database DSN")
	flag.StringVar(&cryptoKey, "crypto-key", "", "Private key path")
	flag.StringVar(&jsonConfigPath, "c", "", "JSON config file path")
	flag.StringVar(&jsonConfigPath, "config", "", "JSON config file path (alias)")
	flag.StringVar(&passwordHashKey, "password-hash-key", "", "Password hash key")
	flag.StringVar(&tokenSignKey, "token-sign-key", "", "Token signing key")
	flag.StringVar(&tokenIssuer, "token-issuer", "", "Token issuer")
	flag.DurationVar(&tokenDuration, "token-duration", 0, "Token duration (e.g., 1h, 30m)")
	flag.DurationVar(&requestTimeout, "request-timeout", 0, "Request timeout (e.g., 30s, 1m)")
	flag.StringVar(&hashKey, "hash-key", "", "Security hash key")

	flag.Parse()

	return &StructuredConfig{
		Services: Services{
			PasswordHashKey: passwordHashKey,
			TokenSignKey:    tokenSignKey,
			TokenIssuer:     tokenIssuer,
			TokenDuration:   tokenDuration,
			HashKey:         hashKey,
		},
		Storage: Storage{
			DB: DB{
				DSN: databaseDSN,
			},
			Files: Files{
				BinaryDataDir: fileStoragePath,
			},
		},
		Server: Server{
			HTTPAddress:    serverAddress.String(),
			GRPCAddress:    grpcServerAddress.String(),
			RequestTimeout: requestTimeout,
		},
		Adapter:      Adapter{},
		Workers:      Workers{},
		JSONFilePath: jsonConfigPath,
	}
}

// String returns a canonical host:port string for a NetAddress.
// If neither Host nor Port are set, it returns the default server address.
func (a *NetAddress) String() string {
	if a.Host == "" && a.Port == 0 {
		return ""
	}

	return a.Host + ":" + strconv.Itoa(a.Port)
}

// Set parses the input string of form host:port and populates the NetAddress.
// It validates the port range, checks IP correctness unless host is "localhost",
// and returns an error if the format or values are invalid.
func (a *NetAddress) Set(s string) error {
	hostAndPort := strings.Split(s, ":")
	if len(hostAndPort) != 2 {
		return errors.New("need address in a form `host:port`")
	}

	host := hostAndPort[0]
	port, err := strconv.Atoi(hostAndPort[1])
	if err != nil {
		return err
	}

	if port < 1 {
		return errors.New("port number is a positive integer")
	}

	if host != "localhost" {
		ip := net.ParseIP(hostAndPort[0])
		if ip == nil {
			return errors.New("incorrect IP-address provided")
		}
	}

	a.Host = host
	a.Port = port
	return nil
}
