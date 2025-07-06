package web_test

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"heimdall/backend/internal/config"
	"heimdall/backend/internal/infra"
	"heimdall/backend/internal/infra/postgres"
	"heimdall/backend/internal/interfaces/web"
	internal_errors "heimdall/backend/pkg/errors"

	"github.com/charmbracelet/lipgloss"
	"github.com/gin-gonic/gin"
)

var (
	testDataStore *infra.DataStore
	testServer    *httptest.Server
	serverLogFile *os.File
)

func TestMain(m *testing.M) {
	// Change directory to the root of the project
	os.Chdir("../../")

	err := setupTestEnvironment()
	if err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	code := m.Run()

	teardownTestEnvironment()
	os.Exit(code)
}

func setupTestEnvironment() error {
	log.Println("🔧 Setting up test environment...")

	TestAppConfig, err := config.LoadConfig(".env.test")
	if err != nil {
		return internal_errors.NewInfrastructureError("Test configuration loading failed", err)
	}

	var Postgres *sql.DB
	Postgres, err = postgres.NewConnection(postgres.Config{
		Host:     TestAppConfig.Database.Postgres.Host,
		Port:     TestAppConfig.Database.Postgres.Port,
		User:     TestAppConfig.Database.Postgres.User,
		Password: TestAppConfig.Database.Postgres.Password,
		DBName:   TestAppConfig.Database.Postgres.DBName,
		SSLMode:  TestAppConfig.Database.Postgres.SSLMode,
	})
	if err != nil {
		return internal_errors.NewInfrastructureError("Postgres connection failed", err)
	}

	log.Println("✅ Postgres connection established")

	// Setup server logging based on environment variable
	err = setupServerLogging()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to setup server logging: %v", err)
	}

	testDataStore = &infra.DataStore{
		Postgres: Postgres,
	}
	testRouter := web.NewRouter(testDataStore)
	testServer = httptest.NewServer(testRouter.Setup())

	log.Printf("🚀 Test server ready at: %s", testServer.URL)
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("")

	return nil
}

func setupServerLogging() error {
	// Always use visual separation mode - show both logs with clear visual distinction
	gin.DefaultWriter = &visualSeparatorWriter{writer: os.Stdout}
	gin.DefaultErrorWriter = &visualSeparatorWriter{writer: os.Stderr}

	return nil
}

type visualSeparatorWriter struct {
	writer io.Writer
}

func (vsw *visualSeparatorWriter) Write(p []byte) (n int, err error) {
	content := strings.TrimRight(string(p), "\n")

	// Parse the log content to extract different parts
	logParts := parseGinLog(content)

	// Create beautiful styled output using lipgloss
	styledOutput := createStyledServerLog(logParts)

	return vsw.writer.Write([]byte(styledOutput))
}

// LogParts represents parsed components of a Gin log entry
type LogParts struct {
	Timestamp    string
	StatusCode   string
	Duration     string
	ClientIP     string
	Method       string
	Path         string
	ErrorMessage string
	IsError      bool
}

// parseGinLog extracts components from Gin log format
func parseGinLog(content string) LogParts {
	// Gin log format: [GIN] 2025/07/18 - 17:25:28 | 200 | 59.371542ms | 127.0.0.1 | POST "/api/v1/auth/login"
	ginLogRegex := regexp.MustCompile(`\[GIN\]\s+(\d{4}/\d{2}/\d{2}\s+-\s+\d{2}:\d{2}:\d{2})\s+\|\s+(\d{3})\s+\|\s+([^|]+)\s+\|\s+([^|]+)\s+\|\s+(\w+)\s+"([^"]+)"`)

	matches := ginLogRegex.FindStringSubmatch(content)
	if len(matches) >= 7 {
		statusCode := matches[2]
		isError := statusCode[0] == '4' || statusCode[0] == '5' // 4xx or 5xx status codes

		return LogParts{
			Timestamp:  matches[1],
			StatusCode: statusCode,
			Duration:   strings.TrimSpace(matches[3]),
			ClientIP:   strings.TrimSpace(matches[4]),
			Method:     matches[5],
			Path:       matches[6],
			IsError:    isError,
		}
	}

	// Check for error messages (usually on next line)
	if strings.Contains(content, "Error #") {
		return LogParts{
			ErrorMessage: content,
			IsError:      true,
		}
	}

	// Fallback for unrecognized format
	return LogParts{
		ErrorMessage: content,
		IsError:      false,
	}
}

// createStyledServerLog creates a beautifully styled server log using lipgloss
func createStyledServerLog(parts LogParts) string {
	// Define modern color palette
	var (
		primaryColor = lipgloss.Color("#6366F1") // Indigo
		successColor = lipgloss.Color("#10B981") // Emerald
		warningColor = lipgloss.Color("#F59E0B") // Amber
		errorColor   = lipgloss.Color("#EF4444") // Red
		mutedColor   = lipgloss.Color("#6B7280") // Gray
		accentColor  = lipgloss.Color("#8B5CF6") // Violet
		infoColor    = lipgloss.Color("#06B6D4") // Cyan
	)

	// Determine the theme based on status
	var (
		headerIcon    string
		headerText    string
		borderColor   lipgloss.Color
		headerBgColor lipgloss.Color
	)

	if parts.IsError {
		headerIcon = "⚠️"
		headerText = "SERVER ERROR"
		borderColor = errorColor
		headerBgColor = errorColor
	} else if parts.StatusCode != "" && parts.StatusCode[0] == '2' {
		headerIcon = "✅"
		headerText = "SERVER SUCCESS"
		borderColor = successColor
		headerBgColor = successColor
	} else if parts.StatusCode != "" && (parts.StatusCode[0] == '4' || parts.StatusCode[0] == '5') {
		headerIcon = "❌"
		headerText = "SERVER ERROR"
		borderColor = errorColor
		headerBgColor = errorColor
	} else {
		headerIcon = "🚀"
		headerText = "SERVER"
		borderColor = primaryColor
		headerBgColor = primaryColor
	}

	// Header style with dynamic theming
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(headerBgColor).
		Bold(true).
		Padding(0, 1).
		MarginTop(1).
		Align(lipgloss.Center)

	// Container style with dynamic border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		MarginBottom(1).
		Width(75) // Fixed width for consistency

	// Content styles with enhanced colors
	timestampStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
	methodStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	pathStyle := lipgloss.NewStyle().Foreground(infoColor).Underline(true)

	// Enhanced status code styling
	var statusStyle lipgloss.Style
	if parts.StatusCode != "" {
		switch parts.StatusCode[0] {
		case '2':
			statusStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true).Background(lipgloss.Color("#DCFCE7")).Padding(0, 1)
		case '3':
			statusStyle = lipgloss.NewStyle().Foreground(infoColor).Bold(true).Background(lipgloss.Color("#E0F2FE")).Padding(0, 1)
		case '4', '5':
			statusStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true).Background(lipgloss.Color("#FEE2E2")).Padding(0, 1)
		default:
			statusStyle = lipgloss.NewStyle().Foreground(warningColor).Bold(true).Background(lipgloss.Color("#FEF3C7")).Padding(0, 1)
		}
	}

	durationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A855F7")).Bold(true) // Purple
	ipStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)

	// Build the content with better formatting
	var content strings.Builder

	if parts.ErrorMessage != "" && parts.StatusCode == "" {
		// This is just an error message
		errorMsgStyle := lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Background(lipgloss.Color("#FEE2E2")).
			Padding(0, 1).
			MarginLeft(1)
		content.WriteString(errorMsgStyle.Render("🔥 " + parts.ErrorMessage))
	} else if parts.StatusCode != "" {
		// This is a full HTTP log entry with enhanced formatting
		content.WriteString(fmt.Sprintf("%s │ %s │ %s │ %s │ %s %s",
			timestampStyle.Render("🕐 "+parts.Timestamp),
			statusStyle.Render(parts.StatusCode),
			durationStyle.Render("⚡ "+parts.Duration),
			ipStyle.Render("🌐 "+parts.ClientIP),
			methodStyle.Render(parts.Method),
			pathStyle.Render(parts.Path),
		))
	} else {
		// Fallback with icon
		content.WriteString("📝 " + parts.ErrorMessage)
	}

	// Create the final styled box with enhanced header
	header := headerStyle.Render(fmt.Sprintf("%s %s", headerIcon, headerText))
	body := containerStyle.Render(content.String())

	// Add subtle shadow effect
	shadow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151")).
		MarginLeft(1).
		Render("▌")

	result := lipgloss.JoinVertical(lipgloss.Left, header, body)
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Top, result, shadow) + "\n"
}
func teardownTestEnvironment() {
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════")
	log.Println("🧹 Tearing down test environment...")

	if testServer != nil {
		testServer.Close()
		log.Print("✅ Test server stopped")
	}

	if testDataStore != nil && testDataStore.Postgres != nil {
		if err := testDataStore.Postgres.Close(); err != nil {
			log.Printf("⚠️ Failed to close Postgres connection: %v", err)
		} else {
			log.Print("✅ Postgres connection closed")
		}
	}

	// Reset Gin writers to default (stdout/stderr)
	gin.DefaultWriter = os.Stdout
	gin.DefaultErrorWriter = os.Stderr

	// Close server log file if it was opened
	if serverLogFile != nil {
		serverLogFile.Close()
		log.Println("📝 Server logs saved to logs/test-server.log")
	}

	log.Println("🏁 Test environment cleanup complete")
}
