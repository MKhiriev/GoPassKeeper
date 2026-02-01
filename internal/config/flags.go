package config

import (
	"errors"
	"flag"
	"net"
	"strconv"
	"strings"
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
func ParseFlags() *StructuredConfig {
	var serverAddress, grpcServerAddress NetAddress
	var fileStoragePath, databaseDSN, privateKeyFilePath string
	var jsonConfigFilePath string

	flag.Var(&serverAddress, "a", "Net address host:port")
	flag.Var(&grpcServerAddress, "grpc-address", "Net grpc server address host:port")
	flag.StringVar(&fileStoragePath, "f", "", "Storage file path string")
	flag.StringVar(&databaseDSN, "d", "", "Postgres database connection string")
	flag.StringVar(&privateKeyFilePath, "crypto-key", "", "Private key file path")

	flag.StringVar(&jsonConfigFilePath, "config", "", "JSON config file path")
	flag.StringVar(&jsonConfigFilePath, "c", "", "JSON config file path")

	flag.Parse()

	return &StructuredConfig{
		// TODO implement me!
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
