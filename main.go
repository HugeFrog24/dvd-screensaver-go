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

// Config holds configurable game parameters
type Config struct {
	MinSpeed  float64 // Minimum speed to avoid stopping
	MaxSpeed  float64 // Maximum speed to avoid chaos
	SpeedStep float64 // Speed adjustment step
}

// LogoRenderer interface for creating logo images
type LogoRenderer interface {
	CreateLogo(width, height int, color color.RGBA) *ebiten.Image
}

// DVDLogoRenderer implements LogoRenderer for DVD-style logos
type DVDLogoRenderer struct{}

// CreateLogo creates a DVD-style logo with the given parameters
func (r *DVDLogoRenderer) CreateLogo(width, height int, color color.RGBA) *ebiten.Image {
	return CreateDVDLogo(width, height, color)
}

// Game implements ebiten.Game interface.
type Game struct {
	logoX             float64
	logoY             float64
	velocityX         float64
	velocityY         float64
	speed             float64 // Current speed multiplier
	logoImage         *ebiten.Image
	logoColors        []color.RGBA
	colorIndex        int
	logger            *log.Logger
	startTime         time.Time
	logBuffer         []string     // Buffer to store recent log messages
	isFullscreen      bool         // Track fullscreen state
	lastFKeyPressed   bool         // Track F key state to detect single presses
	lastEscKeyPressed bool         // Track ESC key state to detect single presses
	lastJKeyPressed   bool         // Track J key state to detect single presses
	lastLKeyPressed   bool         // Track L key state to detect single presses
	config            *Config      // Injected configuration
	logoRenderer      LogoRenderer // Injected logo renderer
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

	// Handle fullscreen controls
	fKeyPressed := ebiten.IsKeyPressed(ebiten.KeyF)
	escKeyPressed := ebiten.IsKeyPressed(ebiten.KeyEscape)

	// F key toggles fullscreen
	if fKeyPressed && !g.lastFKeyPressed {
		g.setFullscreen(!g.isFullscreen, "Fullscreen toggled", elapsed)
	}

	// ESC key exits fullscreen (only when in fullscreen mode)
	if escKeyPressed && !g.lastEscKeyPressed && g.isFullscreen {
		g.setFullscreen(false, "Fullscreen exited with ESC key", elapsed)
	}

	g.lastFKeyPressed = fKeyPressed
	g.lastEscKeyPressed = escKeyPressed

	// Handle speed control with J and L keys
	jKeyPressed := ebiten.IsKeyPressed(ebiten.KeyJ)
	lKeyPressed := ebiten.IsKeyPressed(ebiten.KeyL)

	if jKeyPressed && !g.lastJKeyPressed {
		oldSpeed := g.speed
		g.decreaseSpeed()
		if g.speed != oldSpeed {
			g.logger.Printf("[%s] Speed decreased from %.1f to %.1f",
				elapsed.Round(time.Millisecond), oldSpeed, g.speed)
		}
	}
	g.lastJKeyPressed = jKeyPressed

	if lKeyPressed && !g.lastLKeyPressed {
		oldSpeed := g.speed
		g.increaseSpeed()
		if g.speed != oldSpeed {
			g.logger.Printf("[%s] Speed increased from %.1f to %.1f",
				elapsed.Round(time.Millisecond), oldSpeed, g.speed)
		}
	}
	g.lastLKeyPressed = lKeyPressed

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
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %.1f | Speed: %.1f | [F] fullscreen | [ESC] exit fullscreen | [J] slower | [L] faster", ebiten.CurrentFPS(), g.speed))
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// changeColor changes the logo color
func (g *Game) changeColor() {
	g.colorIndex = (g.colorIndex + 1) % len(g.logoColors)
	currentColor := g.logoColors[g.colorIndex]

	// Create a new logo with the current color using the injected renderer
	g.logoImage = g.logoRenderer.CreateLogo(logoWidth, logoHeight, currentColor)
}

// increaseSpeed increases the logo speed within bounds
func (g *Game) increaseSpeed() {
	oldSpeed := g.speed
	g.speed += g.config.SpeedStep
	if g.speed > g.config.MaxSpeed {
		g.speed = g.config.MaxSpeed
	}

	// Only update velocities if speed actually changed
	if oldSpeed > 0 && g.speed != oldSpeed {
		ratio := g.speed / oldSpeed
		g.velocityX *= ratio
		g.velocityY *= ratio
	}
}

// decreaseSpeed decreases the logo speed within bounds
func (g *Game) decreaseSpeed() {
	oldSpeed := g.speed
	g.speed -= g.config.SpeedStep
	if g.speed < g.config.MinSpeed {
		g.speed = g.config.MinSpeed
	}

	// Only update velocities if speed actually changed
	if oldSpeed > 0 && g.speed != oldSpeed {
		ratio := g.speed / oldSpeed
		g.velocityX *= ratio
		g.velocityY *= ratio
	}
}

// setFullscreen handles fullscreen state changes with logging
func (g *Game) setFullscreen(fullscreen bool, logMessage string, elapsed time.Duration) {
	g.isFullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
	g.logger.Printf("[%s] %s: %v", elapsed.Round(time.Millisecond), logMessage, fullscreen)
}

func main() {
	// Set up random seed
	rand.Seed(time.Now().UnixNano())

	// Create configuration
	config := &Config{
		MinSpeed:  0.5,  // Minimum speed to avoid stopping
		MaxSpeed:  10.0, // Maximum speed to avoid chaos
		SpeedStep: 0.5,  // Speed adjustment step
	}

	// Create logo renderer
	logoRenderer := &DVDLogoRenderer{}

	// Create a new game instance (partially initialized)
	game := &Game{
		logBuffer:         make([]string, 0, maxLogLines),
		startTime:         time.Now(),
		isFullscreen:      false,
		lastFKeyPressed:   false,
		lastEscKeyPressed: false,
		lastJKeyPressed:   false,
		lastLKeyPressed:   false,
		speed:             3.0, // Initial speed
		config:            config,
		logoRenderer:      logoRenderer,
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
	angle := rand.Float64() * 2 * math.Pi
	vx := math.Cos(angle) * game.speed
	vy := math.Sin(angle) * game.speed

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
	logger.Printf("Initial position: (%.2f, %.2f), velocity: (%.2f, %.2f), speed: %.1f, color: %v",
		initialX, initialY, vx, vy, game.speed, game.logoColors[0])

	// Initialize the logo image using the injected renderer
	game.logoImage = game.logoRenderer.CreateLogo(logoWidth, logoHeight, game.logoColors[0])

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
