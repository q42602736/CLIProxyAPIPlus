package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v6/sdk/auth"
	log "github.com/sirupsen/logrus"
)

// DoKiroLogin triggers the Kiro authentication flow with Google OAuth.
// This is the default login method (same as --kiro-google-login).
//
// Parameters:
//   - cfg: The application configuration
//   - options: Login options including Prompt field
func DoKiroLogin(cfg *config.Config, options *LoginOptions) {
	// Use Google login as default
	DoKiroGoogleLogin(cfg, options)
}

// DoKiroGoogleLogin triggers Kiro authentication with Google OAuth.
// This uses a custom protocol handler (kiro://) to receive the callback.
//
// Parameters:
//   - cfg: The application configuration
//   - options: Login options including prompts
func DoKiroGoogleLogin(cfg *config.Config, options *LoginOptions) {
	if options == nil {
		options = &LoginOptions{}
	}

	// Note: Kiro defaults to incognito mode for multi-account support.
	// Users can override with --no-incognito if they want to use existing browser sessions.

	manager := newAuthManager()

	// Use KiroAuthenticator with Google login
	authenticator := sdkAuth.NewKiroAuthenticator()
	record, err := authenticator.LoginWithGoogle(context.Background(), cfg, &sdkAuth.LoginOptions{
		NoBrowser: options.NoBrowser,
		Metadata:  map[string]string{},
		Prompt:    options.Prompt,
	})
	if err != nil {
		log.Errorf("Kiro Google authentication failed: %v", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Make sure the protocol handler is installed")
		fmt.Println("2. Complete the Google login in the browser")
		fmt.Println("3. If callback fails, try: --kiro-import (after logging in via Kiro IDE)")
		return
	}

	// Save the auth record
	savedPath, err := manager.SaveAuth(record, cfg)
	if err != nil {
		log.Errorf("Failed to save auth: %v", err)
		return
	}

	if savedPath != "" {
		fmt.Printf("Authentication saved to %s\n", savedPath)
	}
	if record != nil && record.Label != "" {
		fmt.Printf("Authenticated as %s\n", record.Label)
	}
	fmt.Println("Kiro Google authentication successful!")
}

// DoKiroAWSLogin triggers Kiro authentication with AWS Builder ID.
// This uses the device code flow for AWS SSO OIDC authentication.
//
// Parameters:
//   - cfg: The application configuration
//   - options: Login options including prompts
func DoKiroAWSLogin(cfg *config.Config, options *LoginOptions) {
	if options == nil {
		options = &LoginOptions{}
	}

	// Note: Kiro defaults to incognito mode for multi-account support.
	// Users can override with --no-incognito if they want to use existing browser sessions.

	manager := newAuthManager()

	// Use KiroAuthenticator with AWS Builder ID login (device code flow)
	authenticator := sdkAuth.NewKiroAuthenticator()
	record, err := authenticator.Login(context.Background(), cfg, &sdkAuth.LoginOptions{
		NoBrowser: options.NoBrowser,
		Metadata:  map[string]string{},
		Prompt:    options.Prompt,
	})
	if err != nil {
		log.Errorf("Kiro AWS authentication failed: %v", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Make sure you have an AWS Builder ID")
		fmt.Println("2. Complete the authorization in the browser")
		fmt.Println("3. If callback fails, try: --kiro-import (after logging in via Kiro IDE)")
		return
	}

	// Save the auth record
	savedPath, err := manager.SaveAuth(record, cfg)
	if err != nil {
		log.Errorf("Failed to save auth: %v", err)
		return
	}

	if savedPath != "" {
		fmt.Printf("Authentication saved to %s\n", savedPath)
	}
	if record != nil && record.Label != "" {
		fmt.Printf("Authenticated as %s\n", record.Label)
	}
	fmt.Println("Kiro AWS authentication successful!")
}

// DoKiroAWSAuthCodeLogin triggers Kiro authentication with AWS Builder ID using authorization code flow.
// This provides a better UX than device code flow as it uses automatic browser callback.
//
// Parameters:
//   - cfg: The application configuration
//   - options: Login options including prompts
func DoKiroAWSAuthCodeLogin(cfg *config.Config, options *LoginOptions) {
	if options == nil {
		options = &LoginOptions{}
	}

	// Note: Kiro defaults to incognito mode for multi-account support.
	// Users can override with --no-incognito if they want to use existing browser sessions.

	manager := newAuthManager()

	// Use KiroAuthenticator with AWS Builder ID login (authorization code flow)
	authenticator := sdkAuth.NewKiroAuthenticator()
	record, err := authenticator.LoginWithAuthCode(context.Background(), cfg, &sdkAuth.LoginOptions{
		NoBrowser: options.NoBrowser,
		Metadata:  map[string]string{},
		Prompt:    options.Prompt,
	})
	if err != nil {
		log.Errorf("Kiro AWS authentication (auth code) failed: %v", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Make sure you have an AWS Builder ID")
		fmt.Println("2. Complete the authorization in the browser")
		fmt.Println("3. If callback fails, try: --kiro-aws-login (device code flow)")
		return
	}

	// Save the auth record
	savedPath, err := manager.SaveAuth(record, cfg)
	if err != nil {
		log.Errorf("Failed to save auth: %v", err)
		return
	}

	if savedPath != "" {
		fmt.Printf("Authentication saved to %s\n", savedPath)
	}
	if record != nil && record.Label != "" {
		fmt.Printf("Authenticated as %s\n", record.Label)
	}
	fmt.Println("Kiro AWS authentication successful!")
}

// DoKiroImport imports Kiro token from Kiro IDE's token file.
// This is useful for users who have already logged in via Kiro IDE
// and want to use the same credentials in CLI Proxy API.
//
// Parameters:
//   - cfg: The application configuration
//   - options: Login options (currently unused for import)
func DoKiroImport(cfg *config.Config, options *LoginOptions) {
	if options == nil {
		options = &LoginOptions{}
	}

	manager := newAuthManager()

	// Use ImportFromKiroIDE instead of Login
	authenticator := sdkAuth.NewKiroAuthenticator()
	record, err := authenticator.ImportFromKiroIDE(context.Background(), cfg)
	if err != nil {
		log.Errorf("Kiro token import failed: %v", err)
		fmt.Println("\nMake sure you have logged in to Kiro IDE first:")
		fmt.Println("1. Open Kiro IDE")
		fmt.Println("2. Click 'Sign in with Google' (or GitHub)")
		fmt.Println("3. Complete the login process")
		fmt.Println("4. Run this command again")
		return
	}

	// Save the imported auth record
	savedPath, err := manager.SaveAuth(record, cfg)
	if err != nil {
		log.Errorf("Failed to save auth: %v", err)
		return
	}

	if savedPath != "" {
		fmt.Printf("Authentication saved to %s\n", savedPath)
	}
	if record != nil && record.Label != "" {
		fmt.Printf("Imported as %s\n", record.Label)
	}
	fmt.Println("Kiro token import successful!")
}

// DoKiroImportJSON imports Kiro token via a web page interface.
// Opens a browser with a form to paste JSON, then imports the token.
func DoKiroImportJSON(cfg *config.Config, options *LoginOptions) {
	if options == nil {
		options = &LoginOptions{}
	}

	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Errorf("Failed to start server: %v", err)
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	var jsonData []byte
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})

	mux := http.NewServeMux()

	// Serve the HTML form
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, kiroImportHTMLPage)
	})

	// Handle form submission
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		jsonData = body
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body style="font-family:system-ui;text-align:center;padding:50px;">
			<h2>JSON received! You can close this window.</h2></body></html>`)
		wg.Done()
		close(done)
	})

	server := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Server error: %v", err)
		}
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Printf("Opening browser: %s\n", url)
	fmt.Println("Please paste your JSON in the browser and click Submit.")

	// Open browser
	openBrowser(url)

	// Wait for submission or timeout
	select {
	case <-done:
	case <-time.After(5 * time.Minute):
		fmt.Println("Timeout waiting for JSON input")
		server.Close()
		return
	}

	server.Close()

	if len(jsonData) == 0 {
		log.Error("No JSON data received")
		return
	}

	// Process the JSON
	manager := newAuthManager()
	authenticator := sdkAuth.NewKiroAuthenticator()
	record, err := authenticator.ImportFromJSON(context.Background(), cfg, jsonData)
	if err != nil {
		log.Errorf("Kiro JSON import failed: %v", err)
		return
	}

	savedPath, err := manager.SaveAuth(record, cfg)
	if err != nil {
		log.Errorf("Failed to save auth: %v", err)
		return
	}

	if savedPath != "" {
		fmt.Printf("Authentication saved to %s\n", savedPath)
	}
	if record != nil && record.Label != "" {
		fmt.Printf("Imported as %s\n", record.Label)
	}
	fmt.Println("Kiro JSON import successful!")
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

const kiroImportHTMLPage = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Kiro Token Import</title>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        textarea { width: 100%; height: 300px; font-family: monospace; font-size: 12px; padding: 10px; border: 1px solid #ccc; border-radius: 4px; }
        button { background: #0066cc; color: white; padding: 12px 24px; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; margin-top: 10px; }
        button:hover { background: #0055aa; }
        .hint { color: #666; font-size: 14px; margin-top: 10px; }
    </style>
</head>
<body>
    <h1>Kiro Token Import</h1>
    <p>Paste your Kiro token JSON below:</p>
    <textarea id="json" placeholder='{
  "email": "user@example.com",
  "provider": "BuilderId",
  "accessToken": "aoaAAAAA...",
  "refreshToken": "aorAAAAA...",
  "clientId": "...",
  "clientSecret": "...",
  "region": "us-east-1"
}'></textarea>
    <br>
    <button onclick="submit()">Submit</button>
    <p class="hint">After clicking Submit, you can close this window.</p>
    <script>
        function submit() {
            const json = document.getElementById('json').value;
            if (!json.trim()) { alert('Please paste JSON first'); return; }
            fetch('/submit', { method: 'POST', body: json })
                .then(r => r.text())
                .then(html => { document.body.innerHTML = html; })
                .catch(e => alert('Error: ' + e));
        }
    </script>
</body>
</html>`
