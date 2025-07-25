package commands

import (
    "fmt"
    "io"
    "os"
    
    "github.com/urfave/cli/v2"
    "go.uber.org/zap"
    "gopkg.in/yaml.v3"
    
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

func translateComposeAction(c *cli.Context) error {
    inputFile := c.String("input")
    outputFile := c.String("output")
    
    log := logger.FromContext(c.Context)
    
    // Read input
    var inputData []byte
    var err error
    
    if inputFile == "-" {
        inputData, err = io.ReadAll(os.Stdin)
    } else {
        inputData, err = os.ReadFile(inputFile)
    }
    if err != nil {
        return fmt.Errorf("failed to read input: %w", err)
    }
    
    // Parse runtime spec
    var spec runtime.Spec
    if err := yaml.Unmarshal(inputData, &spec); err != nil {
        return fmt.Errorf("failed to parse runtime spec: %w", err)
    }
    
    log.Info("Translating runtime spec to Docker Compose",
        zap.String("name", spec.Name),
        zap.String("version", spec.Version))
    
    // Generate Docker Compose configuration
    compose := map[string]interface{}{
        "version": "3.8",
        "services": make(map[string]interface{}),
    }
    
    services := compose["services"].(map[string]interface{})
    
    for name, component := range spec.Spec {
        service := map[string]interface{}{
            "image":   fmt.Sprintf("%s@%s", component.Registry, component.Digest),
            "restart": "unless-stopped",
        }
        
        // Add environment variables
        if len(component.Env) > 0 {
            envMap := make(map[string]string)
            for _, env := range component.Env {
                envMap[env.Name] = env.Value
            }
            service["environment"] = envMap
        }
        
        // Add ports
        if len(component.Ports) > 0 {
            var ports []string
            for _, port := range component.Ports {
                ports = append(ports, fmt.Sprintf("%d:%d", port, port))
            }
            service["ports"] = ports
        }
        
        // Add command
        if len(component.Command) > 0 {
            service["command"] = component.Command
        }
        
        services[name] = service
    }
    
    // Write output
    var writer io.Writer
    if outputFile == "-" {
        writer = os.Stdout
    } else {
        file, err := os.Create(outputFile)
        if err != nil {
            return fmt.Errorf("failed to create output file: %w", err)
        }
        defer file.Close()
        writer = file
        log.Info("Writing Docker Compose file", zap.String("file", outputFile))
    }
    
    encoder := yaml.NewEncoder(writer)
    defer encoder.Close()
    
    if err := encoder.Encode(compose); err != nil {
        return fmt.Errorf("failed to write output: %w", err)
    }
    
    log.Info("Successfully translated to Docker Compose")
    return nil
}
