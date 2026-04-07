package collector

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

func NewRemoteFetcher(username, password, command string) func(string) ([]byte, error) {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(password))
	escPassword := buf.String()

	requestFragment := `<modules><clustsafe id="1"/><clustsafe id="2"/><clustsafe id="3"/><clustsafe id="4"/><clustsafe id="5"/><clustsafe id="6"/><clustsafe id="7"/><clustsafe id="8"/></modules>`

	switch command {
	case "sensors":
		requestFragment = "<sensors/>"
	case "all":
		requestFragment += "<sensors/>"
	}

	xmlPayload := fmt.Sprintf(`<clustsafeRequest><authorization><username>%s</username><password>%s</password></authorization>%s</clustsafeRequest>`, username, escPassword, requestFragment)
	reqBuf := []byte(fmt.Sprintf("content-length: %d\n\n%s", len(xmlPayload), xmlPayload))

	conf := &tls.Config{
		InsecureSkipVerify: true,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	return func(host string) ([]byte, error) {
		address := net.JoinHostPort(host, "13000")
		conn, err := tls.Dial("tcp", address, conf)
		if err != nil {
			return nil, fmt.Errorf("tls dial error: %w", err)
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(10 * time.Second))

		if _, err := conn.Write(reqBuf); err != nil {
			return nil, fmt.Errorf("write error: %w", err)
		}

		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header error: %w", err)
		}

		var contentLen int
		if _, err := fmt.Sscanf(strings.TrimSpace(line), "content-length: %d", &contentLen); err != nil {
			return nil, fmt.Errorf("unexpected header or invalid length format: %s", line)
		}

		if _, err := reader.ReadString('\n'); err != nil {
			return nil, fmt.Errorf("failed to read empty line: %w", err)
		}

		respBuf := make([]byte, contentLen)
		if _, err := io.ReadFull(reader, respBuf); err != nil {
			return nil, fmt.Errorf("read body error: %w", err)
		}

		return respBuf, nil
	}
}
