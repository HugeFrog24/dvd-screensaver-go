package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	logoWidth    = 120
	logoHeight   = 60

	// Configuration options
	showLogsInApp = true // Set to true to show logs in the app window
	maxLogLines   = 10   // Maximum number of log lines to show in the app
)

// Game implements ebiten.Game interface.
type Game struct {
	logoX           float64
	logoY           float64
	velocityX       float64
	velocityY       float64
	logoImage       *ebiten.Image
	logoColors      []color.RGBA
	colorIndex      int
	logger          *log.Logger
	startTime       time.Time
	logBuffer       []string // Buffer to store recent log messages
	isFullscreen    bool     // Track fullscreen state
	lastFKeyPressed bool     // Track F key state to detect single presses
}

// LogWriter is a custom io.Writer that captures log messages for display in the app
type LogWriter struct {
	game *Game
}

// Write implements io.Writer interface
func (w *LogWriter) Write(p []byte) (n int, err error) {
	// Add the log message to the game's log buffer
	logMsg := string(p)

	// Only keep the last maxLogLines messages
	if len(w.game.logBuffer) >= maxLogLines {
		// Remove the oldest message
		w.game.logBuffer = w.game.logBuffer[1:]
	}

	// Add the new message
	w.game.logBuffer = append(w.game.logBuffer, logMsg)

	return len(p), nil
}

// Update proceeds the game state.
func (g *Game) Update() error {
	// Update logo position
	g.logoX += g.velocityX
	g.logoY += g.velocityY

	// Get elapsed time since start
	elapsed := time.Since(g.startTime)

	// Handle fullscreen toggle with F key
	fKeyPressed := ebiten.IsKeyPressed(ebiten.KeyF)
	if fKeyPressed && !g.lastFKeyPressed {
		g.isFullscreen = !g.isFullscreen
		ebiten.SetFullscreen(g.isFullscreen)
		g.logger.Printf("[%s] Fullscreen toggled: %v",
			elapsed.Round(time.Millisecond), g.isFullscreen)
	}
	g.lastFKeyPressed = fKeyPressed

	// Check for collision with screen edges
	if g.logoX <= 0 {
		g.velocityX = -g.velocityX
		g.changeColor()
		g.logger.Printf("[%s] BOUNCE: Left edge hit at position (%.2f, %.2f), new velocity: (%.2f, %.2f), new color: %v",
			elapsed.Round(time.Millisecond), g.logoX, g.logoY, g.velocityX, g.velocityY, g.logoColors[g.colorIndex])
	} else if g.logoX+logoWidth >= screenWidth {
		g.velocityX = -g.velocityX
		g.changeColor()
		g.logger.Printf("[%s] BOUNCE: Right edge hit at position (%.2f, %.2f), new velocity: (%.2f, %.2f), new color: %v",
			elapsed.Round(time.Millisecond), g.logoX, g.logoY, g.velocityX, g.velocityY, g.logoColors[g.colorIndex])
	}

	if g.logoY <= 0 {
		g.velocityY = -g.velocityY
		g.changeColor()
		g.logger.Printf("[%s] BOUNCE: Top edge hit at position (%.2f, %.2f), new velocity: (%.2f, %.2f), new color: %v",
			elapsed.Round(time.Millisecond), g.logoX, g.logoY, g.velocityX, g.velocityY, g.logoColors[g.colorIndex])
	} else if g.logoY+logoHeight >= screenHeight {
		g.velocityY = -g.velocityY
		g.changeColor()
		g.logger.Printf("[%s] BOUNCE: Bottom edge hit at position (%.2f, %.2f), new velocity: (%.2f, %.2f), new color: %v",
			elapsed.Round(time.Millisecond), g.logoX, g.logoY, g.velocityX, g.velocityY, g.logoColors[g.colorIndex])
	}

	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw log messages in the background if enabled
	if showLogsInApp && len(g.logBuffer) > 0 {
		// Draw a semi-transparent background for the logs
		logBgColor := color.RGBA{0, 0, 0, 120} // More transparent for background
		vector.DrawFilledRect(screen, 0, float32(screenHeight-20*maxLogLines), float32(screenWidth), float32(20*maxLogLines), logBgColor, false)

		// Draw each log message with a darker color so it doesn't interfere with the logo
		for i, msg := range g.logBuffer {
			// Truncate message if too long
			if len(msg) > 100 {
				msg = msg[:97] + "..."
			}

			// Draw the log message with a darker color
			y := screenHeight - 20*(maxLogLines-i)
			ebitenutil.DebugPrintAt(screen, msg, 10, y)
		}
	}

	// Draw the logo on top of the logs
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.logoX, g.logoY)
	screen.DrawImage(g.logoImage, op)

	// Display FPS and controls info (on top of everything)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %.1f | Press [F] to toggle fullscreen", ebiten.CurrentFPS()))
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// changeColor changes the logo color
func (g *Game) changeColor() {
	g.colorIndex = (g.colorIndex + 1) % len(g.logoColors)
	currentColor := g.logoColors[g.colorIndex]

	// Create a new logo with the current color
	g.logoImage = CreateDVDLogo(logoWidth, logoHeight, currentColor)
}

func main() {
	// Set up random seed
	rand.Seed(time.Now().UnixNano())

	// Create a new game instance (partially initialized)
	game := &Game{
		logBuffer:       make([]string, 0, maxLogLines),
		startTime:       time.Now(),
		isFullscreen:    false,
		lastFKeyPressed: false,
	}

	// Create a custom log writer that will update the game's log buffer
	logWriter := &LogWriter{game: game}

	// Set up logging to file, console, and in-app display
	logFile, err := os.Create("dvd_screensaver.log")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	// Create a multi-writer that writes to file, stdout, and our custom writer
	multiWriter := io.MultiWriter(logFile, os.Stdout, logWriter)

	logger := log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lmicroseconds)

	logger.Printf("DVD Screensaver started at %s", game.startTime.Format(time.RFC3339))

	// Create a new game
	// Generate random velocities with a slight angle
	speed := 3.0
	angle := rand.Float64() * 2 * math.Pi
	vx := math.Cos(angle) * speed
	vy := math.Sin(angle) * speed

	// Ensure we don't have very slow horizontal or vertical movement
	if math.Abs(vx) < 1.0 {
		vx = math.Copysign(1.0, vx)
	}
	if math.Abs(vy) < 1.0 {
		vy = math.Copysign(1.0, vy)
	}

	initialX := float64(rand.Intn(screenWidth - logoWidth))
	initialY := float64(rand.Intn(screenHeight - logoHeight))

	// Complete the game initialization
	game.logoX = initialX
	game.logoY = initialY
	game.velocityX = vx
	game.velocityY = vy
	game.logoColors = []color.RGBA{
		{255, 0, 0, 255},   // Red
		{0, 255, 0, 255},   // Green
		{0, 0, 255, 255},   // Blue
		{255, 255, 0, 255}, // Yellow
		{255, 0, 255, 255}, // Magenta
		{0, 255, 255, 255}, // Cyan
		{255, 165, 0, 255}, // Orange
		{128, 0, 128, 255}, // Purple
	}
	game.logger = logger

	// Log initial state
	logger.Printf("Initial position: (%.2f, %.2f), velocity: (%.2f, %.2f), color: %v",
		initialX, initialY, vx, vy, game.logoColors[0])

	// Initialize the logo image
	game.logoImage = CreateDVDLogo(logoWidth, logoHeight, game.logoColors[0])

	// Set up window
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("DVD Screensaver")

	logger.Printf("Window created with size %dx%d", screenWidth, screenHeight)

	// Run the game
	logger.Printf("Starting game loop")
	if err := ebiten.RunGame(game); err != nil {
		logger.Printf("Game terminated with error: %v", err)
		log.Fatal(err)
	}
}
