package commands

import (
    "fmt"
    "os"
    
    "github.com/olekukonko/tablewriter"
    "github.com/urfave/cli/v2"
    "go.uber.org/zap"
    "gopkg.in/yaml.v3"
    
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func contextListAction(c *cli.Context) error {
    log := logger.FromContext(c.Context)
    
    cfg, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    log.Info("Listing contexts", zap.Int("count", len(cfg.Contexts)))
    
    // Create table
    table := tablewriter.NewWriter(c.App.Writer)
    table.SetHeader([]string{"CURRENT", "NAME", "EXECUTOR ADDRESS", "AVS ADDRESS", "RPC URL"})
    
    for name, ctx := range cfg.Contexts {
        current := ""
        if name == cfg.CurrentContext {
            current = "*"
        }
        
        avsAddr := ctx.AVSAddress
        if avsAddr == "" {
            avsAddr = "-"
        }
        
        rpcURL := ctx.RPCUrl
        if rpcURL == "" {
            rpcURL = "-"
        }
        
        table.Append([]string{
            current,
            name,
            ctx.ExecutorAddress,
            avsAddr,
            rpcURL,
        })
    }
    
    table.Render()
    return nil
}

func contextUseAction(c *cli.Context) error {
    if c.NArg() != 1 {
        return cli.ShowSubcommandHelp(c)
    }
    
    contextName := c.Args().Get(0)
    log := logger.FromContext(c.Context)
    
    cfg, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    if _, exists := cfg.Contexts[contextName]; !exists {
        return fmt.Errorf("context '%s' not found", contextName)
    }
    
    cfg.CurrentContext = contextName
    
    if err := config.SaveConfig(cfg); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }
    
    log.Info("Switched context", zap.String("context", contextName))
    fmt.Printf("Switched to context '%s'\n", contextName)
    
    return nil
}

func contextSetAction(c *cli.Context) error {
    log := logger.FromContext(c.Context)
    
    cfg, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    ctx, exists := cfg.Contexts[cfg.CurrentContext]
    if !exists {
        return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
    }
    
    updated := false
    
    if addr := c.String("executor-address"); addr != "" {
        ctx.ExecutorAddress = addr
        updated = true
        log.Info("Updated executor address", zap.String("address", addr))
    }
    
    if addr := c.String("avs-address"); addr != "" {
        ctx.AVSAddress = addr
        updated = true
        log.Info("Updated AVS address", zap.String("address", addr))
    }
    
    if id := c.Uint64("operator-set-id"); c.IsSet("operator-set-id") {
        ctx.OperatorSetID = uint32(id)
        updated = true
        log.Info("Updated operator set ID", zap.Uint32("id", uint32(id)))
    }
    
    if url := c.String("rpc-url"); url != "" {
        ctx.RPCUrl = url
        updated = true
        log.Info("Updated RPC URL", zap.String("url", url))
    }
    
    if addr := c.String("release-manager"); addr != "" {
        ctx.ReleaseManagerAddress = addr
        updated = true
        log.Info("Updated release manager address", zap.String("address", addr))
    }
    
    if !updated {
        return fmt.Errorf("no values provided to update")
    }
    
    if err := config.SaveConfig(cfg); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }
    
    fmt.Printf("Context '%s' updated\n", cfg.CurrentContext)
    return nil
}

func contextShowAction(c *cli.Context) error {
    log := logger.FromContext(c.Context)
    
    cfg, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    ctx, exists := cfg.Contexts[cfg.CurrentContext]
    if !exists {
        return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
    }
    
    log.Info("Current context", zap.String("name", cfg.CurrentContext))
    
    // Output as YAML
    data := map[string]interface{}{
        "current-context": cfg.CurrentContext,
        "context": map[string]interface{}{
            "executor-address":       ctx.ExecutorAddress,
            "avs-address":           ctx.AVSAddress,
            "operator-set-id":       ctx.OperatorSetID,
            "rpc-url":              ctx.RPCUrl,
            "release-manager":       ctx.ReleaseManagerAddress,
            "network-id":           ctx.NetworkID,
        },
    }
    
    encoder := yaml.NewEncoder(os.Stdout)
    defer encoder.Close()
    
    return encoder.Encode(data)
}
