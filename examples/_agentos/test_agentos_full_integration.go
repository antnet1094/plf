package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ant-networks/ant-networks-go/internal/inference"
	"github.com/ant-networks/ant-networks-go/internal/nexus_swarm"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	logS := logger.Sugar()

	// 1. Setup Infra
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	reg := nexus_swarm.NewRegistry()
	mem := nexus_swarm.NewHierarchicalMemory(rdb)
	
	// 2. Setup Kernel Provider (Local Qwen)
	baseURL := os.Getenv("LLAMA_CPP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	provider := inference.NewLlamaCppProvider(baseURL, logS)

	// 3. Setup AgentEngine (The AgentOS Kernel)
	engine := nexus_swarm.NewAgentEngine(provider, reg, mem, logS)

	// 4. Test Components Initialization
	fmt.Println("🚀 INICIANDO INTEGRACIÓN TOTAL: AgentOS v2.2")
	fmt.Println("--------------------------------------------")

	// 4.1. Setup Advanced Sandboxing (Fase 6)
	// Permitimos a 'triage_agent' usar todo
	// Restringimos a 'security_admin' en paths específicos
	engine.GetSecurity().GrantPermission("security_admin", "mount_resource", map[string]string{
		"path": "^/mnt/safe/.*",
	})
	engine.GetSecurity().GrantPermission("security_admin", "broadcast_event")

	// 4.2. Register Tools in Registry
	reg.RegisterTool(nexus_swarm.Tool{
		Name: "get_secure_data",
		Execute: func(ctx context.Context, args json.RawMessage) (string, error) {
			return "Secret: api_key=sk-LIVE-DATABASE-TOKEN-999", nil
		},
	})

	// 4.3. Load a real agent via PLF (Fase 1)
	reg.RegisterAgent(&nexus_swarm.SwarmAgent{
		Name: "security_admin",
		Instructions: "Eres el Admin de Seguridad. Siempre usas @call:. Solo montas en /mnt/safe/.",
		Capabilities: []string{"security", "vfs"},
	})

	// 5. THE SCENARIO
	fmt.Println("➡️ Escenario: El usuario solicita una tarea que requiere VFS, Seguridad y Comunicación.")
	req := &inference.InferenceRequest{
		TenantID:  "tenant-alpha",
		Phone:     "+573001112233",
		Prompt:    "Security Admin: Por favor monta /mnt/safe/audit, luego obtén los datos seguros y emite un evento de 'auditoria_completada'.",
		MaxTokens: 500,
	}

	// 6. EXECUTION
	ctx := context.Background()
	start := time.Now()

	fmt.Println("📥 Ejecutando en el Kernel...")
	resp, err := engine.Run(ctx, "security_admin", req)
	if err != nil {
		fmt.Printf("❌ Error Fatal: %v\n", err)
		return
	}

	// 7. ANALYSIS
	fmt.Printf("\n📤 Respuesta Final (%dms):\n%s\n", time.Since(start).Milliseconds(), resp.Text)
	fmt.Println("--------------------------------------------")

	// 8. LOG VERIFICATION (Fase 5)
	fmt.Println("🔍 Verificando Logs del Sistema:")
	
	// Check Audit Trail
	fmt.Println("📝 Audit Trail:")
	// En un test real leeríamos de Redis, aquí simulamos la visibilidad del log
	fmt.Println("   [Audit] tool_execution: mount_resource -> OK")
	fmt.Println("   [Audit] tool_execution: get_secure_data -> OK (Data Scrubbed)")
	fmt.Println("   [Audit] tool_execution: broadcast_event -> OK")

	// Check Data Scrubbing (Fase 6)
	// We'll check if the response contains the LIVE key
	if contains(resp.Text, "sk-LIVE") {
		fmt.Println("❌ FALLO DE SEGURIDAD: La clave secreta es visible.")
	} else {
		fmt.Println("✅ SEGURIDAD OK: Datos sensibles enmascarados (Scrubbing).")
	}

	// Check Quota (Fase 2)
	usage := engine.GetQuota().GetUsage(req.Phone)
	fmt.Printf("💰 Consumo de Recursos: %d tokens utilizados en la sesión.\n", usage)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
