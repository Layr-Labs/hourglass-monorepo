package commands

import (
    "fmt"
    "io"
    "os"
    "strings"
    
    "github.com/urfave/cli/v2"
    "go.uber.org/zap"
    "gopkg.in/yaml.v3"
    
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

func translateContainerAction(c *cli.Context) error {
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
    
    log.Info("Translating runtime spec to container commands",
        zap.String("name", spec.Name),
        zap.String("version", spec.Version))
    
    // Generate container run commands
    var commands []string
    commands = append(commands, "#!/bin/bash")
    commands = append(commands, "set -e")
    commands = append(commands, "")
    commands = append(commands, "# Generated from EigenRuntime spec: "+spec.Name)
    commands = append(commands, "# Version: "+spec.Version)
    commands = append(commands, "")
    
    for name, component := range spec.Spec {
        commands = append(commands, fmt.Sprintf("# Component: %s", name))
        
        // Build docker run command
        dockerCmd := []string{"docker", "run", "-d"}
        dockerCmd = append(dockerCmd, "--name", fmt.Sprintf("%s-%s", spec.Name, name))
        dockerCmd = append(dockerCmd, "--restart", "unless-stopped")
        
        // Add environment variables
        for _, env := range component.Env {
            dockerCmd = append(dockerCmd, "-e", fmt.Sprintf("%s=%s", env.Name, env.Value))
        }
        
        // Add ports
        for _, port := range component.Ports {
            dockerCmd = append(dockerCmd, "-p", fmt.Sprintf("%d:%d", port, port))
        }
        
        // Add image
        dockerCmd = append(dockerCmd, fmt.Sprintf("%s@%s", component.Registry, component.Digest))
        
        // Add command
        dockerCmd = append(dockerCmd, component.Command...)
        
        commands = append(commands, strings.Join(dockerCmd, " "))
        commands = append(commands, "")
    }
    
    commands = append(commands, "echo 'All containers started successfully'")
    
    // Write output
    output := strings.Join(commands, "\n")
    
    if outputFile == "-" {
        _, err = os.Stdout.WriteString(output)
    } else {
        err = os.WriteFile(outputFile, []byte(output), 0755)
        if err == nil {
            log.Info("Wrote container script", zap.String("file", outputFile))
        }
    }
    
    if err != nil {
        return fmt.Errorf("failed to write output: %w", err)
    }
    
    log.Info("Successfully translated to container commands")
    return nil
}
