package commands

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

const redirectURI = "https://localhost:8457/callback"

// SetupCmd returns the setup command.
func SetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Authenticate with FreshBooks via OAuth",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}
}

func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to marshal key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func waitForAuthCode() (string, error) {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return "", err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			http.Error(w, "Error: no code received. Close this tab and try again.", http.StatusBadRequest)
			return
		}
		codeCh <- code
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Done! You can close this tab.</h1></body></html>")
	})

	server := &http.Server{
		Addr:    ":8457",
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	ln, err := tls.Listen("tcp", ":8457", server.TLSConfig)
	if err != nil {
		return "", fmt.Errorf("failed to listen on :8457: %w", err)
	}

	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	defer server.Close()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	}
}

func exchangeCodeForToken(clientID, clientSecret, code string) (accessToken, refreshToken string, err error) {
	payload, err := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	})
	if err != nil {
		return "", "", err
	}

	resp, err := http.Post(api.BaseURL+"/auth/oauth/token", "application/json",
		bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, body)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	return result.AccessToken, result.RefreshToken, nil
}

func runSetup() error {
	clientID := os.Getenv("FRESHBOOKS_CLIENT_ID")
	clientSecret := os.Getenv("FRESHBOOKS_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("missing FRESHBOOKS_CLIENT_ID or FRESHBOOKS_CLIENT_SECRET environment variables")
	}

	authURL := fmt.Sprintf("https://auth.freshbooks.com/service/auth/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s",
		url.QueryEscape(clientID), url.QueryEscape(redirectURI))

	fmt.Print("Open this link to authorize freshtime:\n\n")
	fmt.Printf("  %s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	code, err := waitForAuthCode()
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	fmt.Println("Exchanging code for token...")
	accessToken, refreshToken, err := exchangeCodeForToken(clientID, clientSecret, code)
	if err != nil {
		return err
	}

	fmt.Println("Verifying token...")
	httpClient := api.NewHttpClient(accessToken)
	identity, err := api.GetIdentity(httpClient)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	cfg := &config.Config{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccountID:    identity.AccountID,
		BusinessID:   identity.BusinessID,
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("Setup complete.")
	fmt.Printf("  Account:  %s\n", identity.AccountID)
	fmt.Printf("  Business: %d\n", identity.BusinessID)
	fmt.Printf("  Config:   %s\n", config.Path())
	return nil
}
