// Ejemplo de integración PLF en una plataforma Go multi-tenant.
// Muestra cómo cargar, cachear y renderizar archivos .plf
// dentro de un pipeline de agentes WhatsApp.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/antnet1094/plf/pkg/parser"
	"github.com/antnet1094/plf/pkg/renderer"
	"github.com/antnet1094/plf/pkg/types"
	"github.com/antnet1094/plf/pkg/validator"
)

// ─── Agent Registry ───────────────────────────────────────────────────────────

// AgentRegistry carga y cachea documentos PLF por nombre de agente.
// En producción esto vendría de Redis o de un directorio montado.
type AgentRegistry struct {
	mu    sync.RWMutex
	docs  map[string]*types.Document
}

func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{docs: make(map[string]*types.Document)}
}

// Load parsea y valida un .plf, lo almacena con el nombre dado.
func (r *AgentRegistry) Load(name, path string) error {
	doc, err := parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	issues := validator.Validate(doc)
	for _, issue := range issues {
		if issue.Severity == "error" {
			return fmt.Errorf("validation error in %s @%s: %s", name, issue.Section, issue.Message)
		}
		log.Printf("[plf] %s [%s] @%s: %s", name, issue.Severity, issue.Section, issue.Message)
	}
	r.mu.Lock()
	r.docs[name] = doc
	r.mu.Unlock()
	log.Printf("[plf] loaded agent: %s", name)
	return nil
}

// Get devuelve el documento PLF de un agente.
func (r *AgentRegistry) Get(name string) (*types.Document, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	doc, ok := r.docs[name]
	return doc, ok
}

// ─── Message Pipeline ─────────────────────────────────────────────────────────

// IncomingMessage representa un mensaje de WhatsApp entrante.
type IncomingMessage struct {
	TenantID string
	Phone    string
	Body     string
}

// PromptPayload es el resultado listo para enviar a la API del LLM.
type PromptPayload struct {
	AgentName string
	System    string            `json:"system"`
	Messages  []map[string]string `json:"messages"`
}

// Router clasifica el mensaje y produce el payload del agente correcto.
type Router struct {
	registry *AgentRegistry
}

func NewRouter(registry *AgentRegistry) *Router {
	return &Router{registry: registry}
}

// Route construye el prompt para el agente de routing y luego,
// en un pipeline real, enviaría ese prompt al LLM y procesaría
// la respuesta de delegación.
//
// Para este ejemplo devolvemos el payload del router directamente.
func (r *Router) Route(msg IncomingMessage) (*PromptPayload, error) {
	doc, ok := r.registry.Get("whatsapp_router")
	if !ok {
		return nil, fmt.Errorf("agent 'whatsapp_router' not loaded")
	}

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars: map[string]string{
			"tenant_id": msg.TenantID,
			"mensaje":   msg.Body,
		},
		Format: types.FormatNexus,
	})
	if err != nil {
		return nil, fmt.Errorf("render error: %w", err)
	}

	apiPayload := renderer.ToNexus(result)
	
	// Convert NexusMessage to map[string]string for this legacy payload struct
	var msgs []map[string]string
	for _, m := range apiPayload.Messages {
		msgs = append(msgs, map[string]string{
			"role": m.Role,
			"content": m.Content,
		})
	}
	
	payload := &PromptPayload{
		AgentName: "whatsapp_router",
		System:    apiPayload.System,
		Messages:  msgs,
	}
	return payload, nil
}

// BuildAgentPrompt construye el prompt para un agente especializado
// (soporte, ventas, restaurante, etc.) con el mensaje ya clasificado.
func BuildAgentPrompt(
	registry *AgentRegistry,
	agentName string,
	tenantID string,
	userMessage string,
	extraVars map[string]string,
) (*PromptPayload, error) {
	doc, ok := registry.Get(agentName)
	if !ok {
		return nil, fmt.Errorf("agent %q not loaded in registry", agentName)
	}

	vars := map[string]string{
		"mensaje_usuario": userMessage,
		"tenant_id":       tenantID,
	}
	for k, v := range extraVars {
		vars[k] = v
	}

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   vars,
		Format: types.FormatNexus,
	})
	if err != nil {
		return nil, err
	}

	if len(result.UnresolvedVars) > 0 {
		log.Printf("[plf] agent=%s unresolved vars: %v", agentName, result.UnresolvedVars)
	}

	apiPayload := renderer.ToNexus(result)
	
	var msgs []map[string]string
	for _, m := range apiPayload.Messages {
		msgs = append(msgs, map[string]string{
			"role": m.Role,
			"content": m.Content,
		})
	}

	return &PromptPayload{
		AgentName: agentName,
		System:    apiPayload.System,
		Messages:  msgs,
	}, nil
}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	// 1. Inicializar registro y cargar agentes desde disco
	registry := NewAgentRegistry()

	agentsToLoad := map[string]string{
		"whatsapp_router": "examples/whatsapp_router.plf",
		"sysadmin":        "examples/sysadmin.plf",
		"restaurant_bot":  "examples/restaurant_bot.plf",
	}

	for name, path := range agentsToLoad {
		if err := registry.Load(name, path); err != nil {
			log.Printf("[plf] WARN: could not load %s: %v", name, err)
		}
	}

	// 2. Simular mensajes entrantes de distintos tenants
	messages := []IncomingMessage{
		{TenantID: "rest_001", Phone: "+573001234567", Body: "quiero hacer un pedido de bandeja paisa"},
		{TenantID: "tech_002", Phone: "+573009876543", Body: "el servicio PostgreSQL no inicia"},
		{TenantID: "rest_001", Phone: "+573001234568", Body: "cual es el precio del almuerzo"},
	}

	router := NewRouter(registry)

	for _, msg := range messages {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Tenant: %s | Phone: %s\n", msg.TenantID, msg.Phone)
		fmt.Printf("Mensaje: %q\n", msg.Body)

		// 3. Construir prompt de routing
		payload, err := router.Route(msg)
		if err != nil {
			log.Printf("routing error: %v", err)
			continue
		}

		// 4. En producción: enviar payload.System + payload.Messages al LLM,
		//    leer agent_destino del JSON de respuesta, y luego BuildAgentPrompt.
		// Para el ejemplo solo mostramos el payload del router.
		out, _ := json.MarshalIndent(map[string]interface{}{
			"agent":            payload.AgentName,
			"system_len":       len(payload.System),
			"user_message":     payload.Messages[0]["content"],
		}, "", "  ")
		fmt.Println(string(out))
	}

	// 5. Ejemplo directo: construir prompt para el bot de restaurante
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println("Ejemplo: prompt directo para restaurant_bot")

	restaurantPrompt, err := BuildAgentPrompt(
		registry,
		"restaurant_bot",
		"rest_001",
		"tienen domicilios a El Poblado?",
		map[string]string{
			"nombre_restaurante": "La Fogata Paisa",
			"ciudad":             "Medellin",
			"mensaje_cliente":    "tienen domicilios a El Poblado?",
		},
	)
	if err != nil {
		log.Printf("error: %v", err)
	} else {
		fmt.Printf("Agent: %s | System tokens ~%d chars\n",
			restaurantPrompt.AgentName,
			len(restaurantPrompt.System))
		fmt.Printf("User message: %q\n", restaurantPrompt.Messages[0]["content"])
	}
}

