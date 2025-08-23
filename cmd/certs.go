package cmd

import (
    "fmt"
    "path/filepath"

    "github.com/spf13/cobra"
    
    "github.com/sinhaparth5/keynginx/internal/crypto"
    "github.com/sinhaparth5/keynginx/internal/utils"
)

var certsCmd = &cobra.Command{
    Use:   "certs",
    Short: "Generate SSL certificates",
    Long: `Generate SSL private key and self-signed certificate for a domain.

Examples:
  keynginx certs --domain localhost --out ./ssl
  keynginx certs --domain myapp.local --key-size 4096 --validity 730`,
    RunE: runCerts,
}

var (
    certsDomain      string
    certsOutputDir   string
    certsKeySize     int
    certsValidityDays int
    certsOverwrite   bool
    certsCountry     string
    certsState       string
    certsCity        string
    certsOrganization string
    certsUnit        string
    certsEmail       string
)

func init() {
    rootCmd.AddCommand(certsCmd)

    certsCmd.Flags().StringVarP(&certsDomain, "domain", "d", "localhost", "Domain name for certificate (required)")
    certsCmd.Flags().StringVarP(&certsOutputDir, "out", "o", "./ssl", "Output directory for certificates")
    
    certsCmd.Flags().IntVar(&certsKeySize, "key-size", 2048, "RSA key size in bits (2048, 3072, 4096)")
    certsCmd.Flags().IntVar(&certsValidityDays, "validity", 365, "Certificate validity period in days")
    certsCmd.Flags().BoolVar(&certsOverwrite, "overwrite", false, "Overwrite existing certificates")

    certsCmd.Flags().StringVar(&certsCountry, "country", "US", "Country code (2 letters)")
    certsCmd.Flags().StringVar(&certsState, "state", "CA", "State or province")
    certsCmd.Flags().StringVar(&certsCity, "city", "San Francisco", "City or locality")
    certsCmd.Flags().StringVar(&certsOrganization, "organization", "KeyNginx Generated", "Organization name")
    certsCmd.Flags().StringVar(&certsUnit, "unit", "IT Department", "Organizational unit")
    certsCmd.Flags().StringVar(&certsEmail, "email", "", "Email address (optional)")

    certsCmd.MarkFlagRequired("domain")
}

func runCerts(cmd *cobra.Command, args []string) error {
    fmt.Printf("🔐 Generating SSL certificates for domain: %s\n", certsDomain)
    
    if err := validateCertsInput(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    if err := utils.EnsureDirectory(certsOutputDir); err != nil {
        return fmt.Errorf("failed to create output directory: %w", err)
    }

    privateKeyPath := filepath.Join(certsOutputDir, "private.key")
    certificatePath := filepath.Join(certsOutputDir, "certificate.crt")

    if !certsOverwrite && (utils.FileExists(privateKeyPath) || utils.FileExists(certificatePath)) {
        return fmt.Errorf("certificates already exist in %s (use --overwrite to replace)", certsOutputDir)
    }

    generator := crypto.NewGenerator()
    
    certReq := crypto.CertificateRequest{
        Domain:       certsDomain,
        KeySize:      certsKeySize,
        ValidityDays: certsValidityDays,
        Country:      certsCountry,
        State:        certsState,
        City:         certsCity,
        Organization: certsOrganization,
        Unit:         certsUnit,
        Email:        certsEmail,
    }

    verbose := cmd.Flag("verbose").Value.String() == "true"
    if verbose {
        printCertificateDetails(certReq)
    }

    keyPair, err := generator.GenerateKeyPair(certReq)
    if err != nil {
        return fmt.Errorf("certificate generation failed: %w", err)
    }

    if err := generator.SaveKeyPair(keyPair, privateKeyPath, certificatePath); err != nil {
        return fmt.Errorf("failed to save certificates: %w", err)
    }

    fmt.Printf("✅ SSL certificates generated successfully!\n\n")
    fmt.Printf("📁 Output directory: %s\n", certsOutputDir)
    fmt.Printf("🔑 Private key: %s\n", privateKeyPath)
    fmt.Printf("📜 Certificate: %s\n", certificatePath)
    
    if verbose {
        if info, err := generator.ValidateCertificate(certificatePath); err == nil {
            fmt.Printf("\n📋 Certificate Information:\n")
            fmt.Printf("   Subject: %s\n", info.Subject)
            fmt.Printf("   Valid from: %s\n", info.NotBefore.Format("2006-01-02 15:04:05"))
            fmt.Printf("   Valid until: %s\n", info.NotAfter.Format("2006-01-02 15:04:05"))
            fmt.Printf("   Days remaining: %d\n", info.DaysUntilExpiry)
            if len(info.DNSNames) > 1 {
                fmt.Printf("   DNS names: %v\n", info.DNSNames)
            }
        }
    }
    
    fmt.Printf("\n💡 Next steps:\n")
    fmt.Printf("   • Use these certificates in your web server configuration\n")
    fmt.Printf("   • For nginx: ssl_certificate %s; ssl_certificate_key %s;\n", 
        certificatePath, privateKeyPath)
    
    return nil
}

func validateCertsInput() error {
    if certsDomain == "" {
        return fmt.Errorf("domain is required")
    }

    validKeySizes := map[int]bool{2048: true, 3072: true, 4096: true}
    if !validKeySizes[certsKeySize] {
        return fmt.Errorf("invalid key size %d (must be 2048, 3072, or 4096)", certsKeySize)
    }

    if certsValidityDays <= 0 {
        return fmt.Errorf("validity days must be positive (got %d)", certsValidityDays)
    }

    if certsValidityDays > 3650 {
        return fmt.Errorf("validity days too high %d (maximum 3650 = 10 years)", certsValidityDays)
    }

    return nil
}

func printCertificateDetails(req crypto.CertificateRequest) {
    fmt.Printf("\n📝 Certificate Details:\n")
    fmt.Printf("   Domain: %s\n", req.Domain)
    fmt.Printf("   Key Size: %d bits\n", req.KeySize)
    fmt.Printf("   Validity: %d days\n", req.ValidityDays)
    fmt.Printf("   Country: %s\n", req.Country)
    fmt.Printf("   State: %s\n", req.State)
    fmt.Printf("   City: %s\n", req.City)
    fmt.Printf("   Organization: %s\n", req.Organization)
    if req.Email != "" {
        fmt.Printf("   Email: %s\n", req.Email)
    }
    fmt.Println()
}