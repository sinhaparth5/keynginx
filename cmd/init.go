package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sinhaparth5/keynginx/internal/config"
	"github.com/sinhaparth5/keynginx/internal/crypto"
	"github.com/sinhaparth5/keynginx/internal/nginx"
	"github.com/sinhaparth5/keynginx/internal/utils"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new KeyNginx project",
	Long: `Initialize creates a complete KeyNginx project with:
- SSL certificates
- Nginx configuration file with security headers
- Docker Compose configuration
- Project configuration file

This sets up everything needed for a secure web server.`,
	RunE: runInit,
}

var (
	initDomain        string
	initOutputDir     string
	initInteractive   bool
	initSecurityLevel string
	initHTTPSPort     int
	initHTTPPort      int
	initOverwrite     bool
	initServices      []string
	initCustomHeaders []string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initDomain, "domain", "d", "localhost", "Domain name for the project")
	initCmd.Flags().StringVarP(&initOutputDir, "output", "o", "./keynginx-output", "Output directory")
	initCmd.Flags().BoolVarP(&initInteractive, "interactive", "i", false, "Run in interactive mode")
	initCmd.Flags().StringVar(&initSecurityLevel, "security-level", "balanced", "Security level (strict/balanced/permissive)")
	initCmd.Flags().IntVar(&initHTTPSPort, "https-port", 8443, "HTTPS port")
	initCmd.Flags().IntVar(&initHTTPPort, "http-port", 8080, "HTTP port")
	initCmd.Flags().BoolVar(&initOverwrite, "overwrite", false, "Overwrite existing files")
	initCmd.Flags().StringSliceVar(&initServices, "services", []string{}, "Services in format 'name:port:path' (e.g. 'frontend:3000:/')")
	initCmd.Flags().StringSliceVar(&initCustomHeaders, "custom-headers", []string{}, "Custom headers in format 'Key:Value'")
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 KeyNginx Project Initialization")
	fmt.Println("===================================")

	cfg := config.NewDefaultConfig()

	if initInteractive {
		if err := runInteractiveInit(cfg); err != nil {
			return fmt.Errorf("interactive initialization failed: %w", err)
		}
	} else {
		configureFromFlags(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if utils.DirectoryExists(cfg.Project.OutputDir) && !initOverwrite {
		return fmt.Errorf("output directory %s already exists (use --overwrite)", cfg.Project.OutputDir)
	}

	fmt.Printf("📁 Creating project in %s...\n", cfg.Project.OutputDir)
	if err := utils.EnsureDirectory(cfg.Project.OutputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("🔐 Generating SSL certificates for %s...\n", cfg.Project.Domain)
	if err := generateSSLCertificates(cfg); err != nil {
		return fmt.Errorf("failed to generate SSL certificates: %w", err)
	}

	fmt.Println("⚙️  Generating Nginx configuration...")
	if err := generateNginxConfiguration(cfg); err != nil {
		return fmt.Errorf("failed to generate Nginx configuration: %w", err)
	}

	fmt.Println("🐳 Generating Docker Compose configuration...")
	if err := generateDockerCompose(cfg); err != nil {
		return fmt.Errorf("failed to generate Docker Compose: %w", err)
	}

	fmt.Println("💾 Saving project configuration...")
	configFile := filepath.Join(cfg.Project.OutputDir, "keynginx.yaml")
	if err := cfg.Save(configFile); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	printInitSuccess(cfg)

	return nil
}

func runInteractiveInit(cfg *config.Config) error {
	fmt.Println("\n🎯 Interactive Project Setup")

	fmt.Printf("Domain name (default: %s): ", cfg.Project.Domain)
	var domain string
	fmt.Scanln(&domain)
	if domain != "" {
		cfg.Project.Domain = domain
		cfg.Nginx.ServerName = domain
	}

	fmt.Printf("Security level [strict/balanced/permissive] (default: %s): ", cfg.Security.Level)
	var secLevel string
	fmt.Scanln(&secLevel)
	if secLevel != "" {
		cfg.SetSecurityLevel(secLevel)
	}

	fmt.Print("Add services? [y/N]: ")
	var addServices string
	fmt.Scanln(&addServices)
	if strings.ToLower(addServices) == "y" {
		return addInteractiveServices(cfg)
	}

	return nil
}

func addInteractiveServices(cfg *config.Config) error {
	for {
		fmt.Print("Service name (or 'done' to finish): ")
		var name string
		fmt.Scanln(&name)

		if name == "done" || name == "" {
			break
		}

		fmt.Printf("Port for %s: ", name)
		var port int
		fmt.Scanln(&port)

		fmt.Printf("Path for %s (default: /): ", name)
		var path string
		fmt.Scanln(&path)
		if path == "" {
			path = "/"
		}

		cfg.AddService(name, port, path)
		fmt.Printf("✅ Added service: %s -> %s:%d%s\n", name, name, port, path)
	}

	return nil
}

func configureFromFlags(cfg *config.Config) {
	cfg.Project.Domain = initDomain
	cfg.Project.OutputDir = initOutputDir
	cfg.Nginx.ServerName = initDomain
	cfg.Nginx.HTTPSPort = initHTTPSPort
	cfg.Nginx.HTTPPort = initHTTPPort

	cfg.SetSecurityLevel(initSecurityLevel)

	for _, service := range initServices {
		parts := strings.Split(service, ":")
		if len(parts) == 3 {
			name := parts[0]
			port := 0
			fmt.Sscanf(parts[1], "%d", &port)
			path := parts[2]
			cfg.AddService(name, port, path)
		}
	}

	for _, header := range initCustomHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			cfg.Nginx.CustomHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
}

func generateSSLCertificates(cfg *config.Config) error {
	generator := crypto.NewGenerator()

	certReq := crypto.CertificateRequest{
		Domain:       cfg.Project.Domain,
		KeySize:      cfg.SSL.KeySize,
		ValidityDays: cfg.SSL.ValidityDays,
		Country:      cfg.SSL.Country,
		State:        cfg.SSL.State,
		City:         cfg.SSL.City,
		Organization: cfg.SSL.Organization,
		Unit:         cfg.SSL.Unit,
		Email:        cfg.SSL.Email,
	}

	keyPair, err := generator.GenerateKeyPair(certReq)
	if err != nil {
		return err
	}

	sslDir := filepath.Join(cfg.Project.OutputDir, "ssl")
	if err := utils.EnsureDirectory(sslDir); err != nil {
		return err
	}

	privateKeyPath := filepath.Join(sslDir, "private.key")
	certificatePath := filepath.Join(sslDir, "certificate.crt")

	return generator.SaveKeyPair(keyPair, privateKeyPath, certificatePath)
}

func generateNginxConfiguration(cfg *config.Config) error {
	generator := nginx.NewGenerator()

	nginxConfig, err := generator.GenerateConfig(cfg)
	if err != nil {
		return err
	}

	nginxConfigPath := filepath.Join(cfg.Project.OutputDir, "nginx.conf")
	return os.WriteFile(nginxConfigPath, []byte(nginxConfig), 0644)
}

func generateDockerCompose(cfg *config.Config) error {
	generator := nginx.NewGenerator()

	dockerConfig, err := generator.GenerateDockerCompose(cfg)
	if err != nil {
		return err
	}

	dockerComposePath := filepath.Join(cfg.Project.OutputDir, "docker-compose.yml")
	return os.WriteFile(dockerComposePath, []byte(dockerConfig), 0644)
}

func printInitSuccess(cfg *config.Config) {
	fmt.Println("\n🎉 KeyNginx Project Created Successfully!")
	fmt.Println("========================================")

	fmt.Printf("📁 Project: %s\n", cfg.Project.OutputDir)
	fmt.Printf("🌐 Domain: %s\n", cfg.Project.Domain)
	fmt.Printf("🔐 HTTPS: :%d\n", cfg.Nginx.HTTPSPort)
	fmt.Printf("🛡️  Security: %s\n", cfg.Security.Level)

	if len(cfg.Nginx.Services) > 0 {
		fmt.Println("\n🔄 Services:")
		for _, service := range cfg.Nginx.Services {
			fmt.Printf("   • %s: %s -> %s:%d\n", service.Name, service.Path, service.Name, service.Port)
		}
	}

	fmt.Println("\n📋 Generated Files:")
	fmt.Printf("   • ssl/private.key (SSL private key)\n")
	fmt.Printf("   • ssl/certificate.crt (SSL certificate)\n")
	fmt.Printf("   • nginx.conf (Nginx configuration)\n")
	fmt.Printf("   • docker-compose.yml (Docker setup)\n")
	fmt.Printf("   • keynginx.yaml (Project configuration)\n")

	fmt.Println("\n🚀 Next Steps:")
	fmt.Printf("   1. cd %s\n", cfg.Project.OutputDir)
	fmt.Printf("   2. docker-compose up -d\n")
	fmt.Printf("   3. Visit https://localhost:%d\n", cfg.Nginx.HTTPSPort)

	if cfg.Project.Domain != "localhost" {
		fmt.Printf("\n💡 For domain '%s', add to /etc/hosts:\n", cfg.Project.Domain)
		fmt.Printf("   127.0.0.1 %s\n", cfg.Project.Domain)
	}

	fmt.Println("\n🔧 Customize your setup:")
	fmt.Println("   • Edit nginx.conf for advanced configuration")
	fmt.Println("   • Modify docker-compose.yml to add your services")
	fmt.Println("   • Update keynginx.yaml and regenerate with 'keynginx init'")
}
