package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/smgarciagr/codeact-agent-drama/internal/agent"
	"github.com/smgarciagr/codeact-agent-drama/internal/database"
	"github.com/smgarciagr/codeact-agent-drama/internal/models"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the database and insert initial data
	database.InitDB()

	// Initialize the AI Client based on LLM_PROVIDER env var
	// Options: "ollama" (default, local) or "groq" (cloud, free API key)
	var agentClient agent.LLMClient
	provider := os.Getenv("LLM_PROVIDER")
	switch provider {
	case "groq":
		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			log.Fatal("GROQ_API_KEY is required when LLM_PROVIDER=groq")
		}
		agentClient = agent.NewGroqClient(apiKey, os.Getenv("GROQ_MODEL"))
		log.Println("Using Groq (cloud) as LLM provider")
	default:
		agentClient = agent.NewOllamaClient(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))
		log.Println("Using Ollama (local) as LLM provider")
	}

	// Configure the web server
	app := fiber.New()

	// Serve static files from the "static" directory
	app.Static("/", "./static")

	// Serve exported files for download
	app.Get("/exports/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		return c.Download("./"+filename, filename)
	})

	// Endpoint to list dramas
	app.Get("/api/dramas", func(c *fiber.Ctx) error {
		var dramas []models.Drama
		database.DB.Find(&dramas)
		return c.JSON(dramas)
	})

	// Endpoint for the Agent Command (CodeAct logic)
	app.Post("/api/agent/command", func(c *fiber.Ctx) error {
		type Request struct {
			Command string `json:"command"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// CodeAct loop: generate code, execute, retry with feedback if it fails
		const maxRetries = 3
		var action *agent.AgentAction
		var output string
		var attempts int
		currentCommand := req.Command

		for attempt := 1; attempt <= maxRetries; attempt++ {
			attempts = attempt
			// 1. Generate the action (Thought + Code) using AI
			var genErr error
			action, genErr = agentClient.GenerateAction(context.Background(), currentCommand)
			if genErr != nil {
				log.Printf("LLM API error (attempt %d): %v", attempt, genErr)
				if attempt == maxRetries {
					return c.Status(500).JSON(fiber.Map{
						"thought": "Error calling AI",
						"error":   "Failed to generate agent action: " + genErr.Error(),
					})
				}
				currentCommand = fmt.Sprintf("%s\n\nPrevious attempt failed with error: %s\nPlease fix the issue and try again.", req.Command, genErr.Error())
				continue
			}

			// 2. Execute the generated Go code
			var execErr error
			output, execErr = agent.ExecuteCode(action.Code)

			// 3. Log every attempt for audit
			logFile, _ := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			logFile.WriteString(fmt.Sprintf("Attempt: %d\nThought: %s\nOutput: %s\n---\n", attempt, action.Thought, output))
			logFile.Close()

			if execErr == nil {
				// Success — break out of the retry loop
				break
			}

			log.Printf("Execution error (attempt %d): %v", attempt, execErr)
			if attempt == maxRetries {
				return c.Status(500).JSON(fiber.Map{
					"thought": action.Thought,
					"error":   "Execution error after retries: " + execErr.Error(),
					"output":  output,
				})
			}

			// Feedback loop: send the error back to the AI so it can fix its code
			currentCommand = fmt.Sprintf("%s\n\nYour previous code produced this error:\n%s\nOutput: %s\nPlease fix the code and try again.", req.Command, execErr.Error(), output)
		}

		// Return the agent's thought and the execution result
		return c.JSON(fiber.Map{
			"thought":  action.Thought,
			"code":     action.Code,
			"output":   output,
			"attempts": attempts,
			"success":  true,
		})
	})

	// Start the server on port 3000
	log.Fatal(app.Listen(":3000"))
}
