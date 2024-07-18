package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// Get 获取TLS配置
func Get(certFile, keyFile, caFile, serverName string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		ServerName: serverName,
	}

	// 如果提供了证书文件和私钥文件，则加载它们
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("加载证书和私钥失败: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// 如果提供了CA证书文件，则加载它
	if caFile != "" {
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("读取CA证书失败: %v", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	} else {
		tlsConfig.InsecureSkipVerify = true // 没有CA证书则跳过验证
	}

	return tlsConfig, nil
}
